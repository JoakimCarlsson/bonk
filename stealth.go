package bonk

import (
	"regexp"
	"runtime"
	"strings"

	"github.com/joakimcarlsson/bonk/proto"
)

var chromeVersionRe = regexp.MustCompile(`Chrome/(\d+)\.(\d+\.\d+\.\d+)`)

func applyStealth(p *Page) error {
	if err := patchUserAgent(p); err != nil {
		return err
	}
	if err := proto.EmulationSetHardwareConcurrencyOverride(8).
		Do(p.execCtx); err != nil {
		return err
	}
	return nil
}

func patchUserAgent(p *Page) error {
	res, err := proto.RuntimeEvaluate("navigator.userAgent").
		WithReturnByValue(true).
		Do(p.execCtx)
	if err != nil {
		return err
	}

	ua := strings.Trim(string(res.Result.Value), `"`)
	cleanUA := strings.ReplaceAll(ua, "Headless", "")

	major, full := extractChromeVersion(cleanUA)
	if major == "" {
		major = "136"
		full = "136.0.0.0"
	}

	platform := detectPlatform()

	return proto.EmulationSetUserAgentOverride(cleanUA).
		WithAcceptLanguage("en-US,en;q=0.9").
		WithPlatform(platform).
		WithUserAgentMetadata(proto.EmulationUserAgentMetadata{
			Brands: []proto.EmulationUserAgentBrandVersion{
				{Brand: "Chromium", Version: major},
				{Brand: "Google Chrome", Version: major},
				{Brand: "Not_A Brand", Version: "8"},
			},
			FullVersionList: []proto.EmulationUserAgentBrandVersion{
				{Brand: "Chromium", Version: full},
				{Brand: "Google Chrome", Version: full},
				{Brand: "Not_A Brand", Version: "8.0.0.0"},
			},
			Platform:        platform,
			PlatformVersion: "",
			Architecture:    detectArch(),
			Model:           "",
			Mobile:          false,
		}).
		Do(p.execCtx)
}

func injectPatches(p *Page) error {
	_, err := proto.PageAddScriptToEvaluateOnNewDocument(stealthPatchScript).
		Do(p.execCtx)
	return err
}

func extractChromeVersion(ua string) (major, full string) {
	matches := chromeVersionRe.FindStringSubmatch(ua)
	if len(matches) < 3 {
		return "", ""
	}
	return matches[1], matches[1] + "." + matches[2]
}

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

func detectArch() string {
	switch runtime.GOARCH {
	case "arm64", "aarch64":
		return "arm"
	default:
		return "x86"
	}
}
