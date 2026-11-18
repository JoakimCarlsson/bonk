package bonk

import (
	"encoding/json"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/joakimcarlsson/bonk/proto"
)

var chromeVersionRe = regexp.MustCompile(`Chrome/(\d+)\.(\d+\.\d+\.\d+)`)

// greaseChars are the characters Chromium picks from to build its GREASE
// brand name.
var greaseChars = []string{" ", "(", ":", "-", ".", "/", ")", ";", "=", "?", "_"}

// greaseVersions are the versions Chromium picks from for its GREASE brand.
var greaseVersions = []string{"8", "99", "24"}

// greaseOrders are the brand orderings Chromium permutes between, indexed by
// the major version modulo six; each lists the positions of the GREASE
// brand, the Chromium brand and the product brand.
var greaseOrders = [][3]int{
	{0, 1, 2}, {0, 2, 1}, {1, 0, 2}, {1, 2, 0}, {2, 0, 1}, {2, 1, 0},
}

// applyStealth makes a freshly attached page look like a regular desktop
// Chrome. Only headless Chrome gives itself away in its user agent and
// client hints, so a headful browser keeps its native identity untouched.
// It returns the error of the first CDP command that fails.
func applyStealth(p *Page) error {
	return patchUserAgent(p)
}

// patchUserAgent replaces the "HeadlessChrome" user agent and client hints
// of a headless browser with those a headful Google Chrome of the same
// version sends, with the brand list and GREASE brand Chromium itself
// derives from the major version. A headful browser is left as it is. It
// returns an error when the user agent cannot be read or overridden.
func patchUserAgent(p *Page) error {
	res, err := proto.RuntimeEvaluate("navigator.userAgent").
		WithReturnByValue(true).
		Do(p.execCtx)
	if err != nil {
		return err
	}

	var ua string
	if json.Unmarshal(res.Result.Value, &ua) != nil {
		ua = strings.Trim(string(res.Result.Value), `"`)
	}
	if !strings.Contains(ua, "Headless") {
		return nil
	}
	cleanUA := strings.ReplaceAll(ua, "HeadlessChrome", "Chrome")

	major, full := extractChromeVersion(cleanUA)
	if v, err := proto.BrowserGetVersion().Do(p.execCtx); err == nil {
		if m, f := extractChromeVersion(
			strings.ReplaceAll(v.Product, "HeadlessChrome", "Chrome"),
		); m != "" {
			major, full = m, f
		}
	}
	if major == "" {
		major = "136"
		full = "136.0.0.0"
	}

	platform := detectPlatform()
	product := "Google Chrome"
	if p.browserCtx.browser.cfg.chromium {
		product = ""
	}
	brands, fullList := clientHintBrands(major, full, product)

	return proto.EmulationSetUserAgentOverride(cleanUA).
		WithPlatform(platformNavigator()).
		WithUserAgentMetadata(proto.EmulationUserAgentMetadata{
			Brands:          brands,
			FullVersionList: fullList,
			FullVersion:     full,
			Platform:        platform,
			PlatformVersion: detectPlatformVersion(),
			Architecture:    detectArch(),
			Bitness:         "64",
			Model:           "",
			Mobile:          false,
			Wow64:           false,
		}).
		Do(p.execCtx)
}

// clientHintBrands builds the brand list and full version list a headful
// browser of the given major and full version reports, ordered and greased
// the way Chromium's GetUserAgentBrandList does it. product is the product
// brand, "Google Chrome", or empty for a plain Chromium, which reports only
// the Chromium and GREASE brands.
func clientHintBrands(
	major, full, product string,
) (brands, fullList []proto.EmulationUserAgentBrandVersion) {
	seed, _ := strconv.Atoi(major)
	grease := "Not" + greaseChars[seed%len(greaseChars)] + "A" +
		greaseChars[(seed+1)%len(greaseChars)] + "Brand"
	greaseVersion := greaseVersions[seed%len(greaseVersions)]

	names := []string{grease, "Chromium"}
	short := []string{greaseVersion, major}
	long := []string{greaseVersion + ".0.0.0", full}
	order := []int{seed % 2, (seed + 1) % 2}
	if product != "" {
		names = append(names, product)
		short = append(short, major)
		long = append(long, full)
		o := greaseOrders[seed%len(greaseOrders)]
		order = o[:]
	}

	brands = make([]proto.EmulationUserAgentBrandVersion, len(names))
	fullList = make([]proto.EmulationUserAgentBrandVersion, len(names))
	for i, name := range names {
		brands[order[i]] = proto.EmulationUserAgentBrandVersion{
			Brand: name, Version: short[i],
		}
		fullList[order[i]] = proto.EmulationUserAgentBrandVersion{
			Brand: name, Version: long[i],
		}
	}
	return brands, fullList
}

// extractChromeVersion returns the major and full Chrome version a user
// agent or product string carries, or empty strings when it has none.
func extractChromeVersion(ua string) (major, full string) {
	matches := chromeVersionRe.FindStringSubmatch(ua)
	if len(matches) < 3 {
		return "", ""
	}
	return matches[1], matches[1] + "." + matches[2]
}

// detectPlatform returns the client hints platform name of the host OS.
func detectPlatform() string {
	switch runtime.GOOS {
	case "darwin":
		return "macOS"
	case "windows":
		return "Windows"
	default:
		return "Linux"
	}
}

// platformNavigator returns the navigator.platform value Chrome reports on
// the host OS.
func platformNavigator() string {
	switch runtime.GOOS {
	case "darwin":
		return "MacIntel"
	case "windows":
		return "Win32"
	default:
		if runtime.GOARCH == "arm64" {
			return "Linux aarch64"
		}
		return "Linux x86_64"
	}
}

// detectPlatformVersion returns a plausible client hints platform version
// for the host OS; Chrome on Linux reports an empty one.
func detectPlatformVersion() string {
	switch runtime.GOOS {
	case "darwin":
		return "15.0.0"
	case "windows":
		return "19.0.0"
	default:
		return ""
	}
}

// detectArch returns the client hints architecture of the host CPU.
func detectArch() string {
	switch runtime.GOARCH {
	case "arm64", "aarch64":
		return "arm"
	default:
		return "x86"
	}
}
