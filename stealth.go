package bonk

import (
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/joakimcarlsson/bonk/proto"
)

var (
	chromeVersionRe = regexp.MustCompile(`Chrome/(\d+)\.(\d+\.\d+\.\d+)`)
	edgeVersionRe   = regexp.MustCompile(`Edg/(\d+)\.(\d+\.\d+\.\d+)`)
)

func applyStealth(p *Page, locale string) error {
	if err := patchUserAgent(p, locale); err != nil {
		return err
	}
	if err := addStealthScript(p, locale); err != nil {
		return err
	}
	if err := proto.EmulationSetAutomationOverride(false).
		Do(p.execCtx); err != nil {
		return err
	}
	if err := proto.EmulationSetHardwareConcurrencyOverride(8).
		Do(p.execCtx); err != nil {
		return err
	}
	return nil
}

func patchUserAgent(p *Page, locale string) error {
	res, err := proto.RuntimeEvaluate("navigator.userAgent").
		WithReturnByValue(true).
		Do(p.execCtx)
	if err != nil {
		return err
	}

	ua := strings.Trim(string(res.Result.Value), `"`)
	cleanUA := strings.ReplaceAll(ua, "Headless", "")

	major, full := extractChromeVersion(cleanUA)
	brand := "Google Chrome"
	if edgeMajor, edgeFull := extractEdgeVersion(cleanUA); edgeMajor != "" {
		major = edgeMajor
		full = edgeFull
		brand = "Microsoft Edge"
	}
	if major == "" {
		major = "136"
		full = "136.0.0.0"
	}

	platform := detectPlatform()
	platformVersion := detectPlatformVersion()
	acceptLanguage := acceptLanguageForLocale(locale)

	return proto.EmulationSetUserAgentOverride(cleanUA).
		WithAcceptLanguage(acceptLanguage).
		WithPlatform(platform).
		WithUserAgentMetadata(proto.EmulationUserAgentMetadata{
			Brands: []proto.EmulationUserAgentBrandVersion{
				{Brand: "Chromium", Version: major},
				{Brand: brand, Version: major},
				{Brand: "Not_A Brand", Version: "8"},
			},
			FullVersionList: []proto.EmulationUserAgentBrandVersion{
				{Brand: "Chromium", Version: full},
				{Brand: brand, Version: full},
				{Brand: "Not_A Brand", Version: "8.0.0.0"},
			},
			Platform:        platform,
			PlatformVersion: platformVersion,
			Architecture:    detectArch(),
			Model:           "",
			Mobile:          false,
		}).
		Do(p.execCtx)
}

func addStealthScript(p *Page, locale string) error {
	language := locale
	if language == "" {
		language = "en-US"
	}
	baseLanguage := language
	if i := strings.IndexByte(baseLanguage, '-'); i > 0 {
		baseLanguage = baseLanguage[:i]
	}
	return p.AddInitScript(`
const defineGetter = (target, prop, getter) => {
	try {
		Object.defineProperty(target, prop, { get: getter, configurable: true });
	} catch (_) {}
};
defineGetter(Navigator.prototype, 'webdriver', () => undefined);
defineGetter(Navigator.prototype, 'language', () => ` + quoteJS(language) + `);
defineGetter(Navigator.prototype, 'languages', () => [` + quoteJS(language) + `, ` + quoteJS(baseLanguage) + `]);
defineGetter(Navigator.prototype, 'vendor', () => 'Google Inc.');
defineGetter(Navigator.prototype, 'deviceMemory', () => 8);
defineGetter(Navigator.prototype, 'plugins', () => [
	{name: 'PDF Viewer'},
	{name: 'Chrome PDF Viewer'},
	{name: 'Chromium PDF Viewer'},
	{name: 'Microsoft Edge PDF Viewer'},
	{name: 'WebKit built-in PDF'},
]);
Object.defineProperty(navigator, 'plugins', {
	get: () => [
		{name: 'PDF Viewer'},
		{name: 'Chrome PDF Viewer'},
		{name: 'Chromium PDF Viewer'},
	],
	configurable: true,
});
window.chrome = window.chrome || {};
window.chrome.app = window.chrome.app || {
	isInstalled: false,
	InstallState: { DISABLED: 'disabled', INSTALLED: 'installed', NOT_INSTALLED: 'not_installed' },
	RunningState: { CANNOT_RUN: 'cannot_run', READY_TO_RUN: 'ready_to_run', RUNNING: 'running' },
};
window.chrome.runtime = window.chrome.runtime || {};
defineGetter(Screen.prototype, 'availTop', () => 0);
defineGetter(Screen.prototype, 'availLeft', () => 0);
defineGetter(Screen.prototype, 'availWidth', () => window.innerWidth || 1365);
defineGetter(Screen.prototype, 'availHeight', () => window.innerHeight || 768);
defineGetter(Screen.prototype, 'width', () => window.innerWidth || 1365);
defineGetter(Screen.prototype, 'height', () => window.innerHeight || 768);
defineGetter(window, 'outerWidth', () => window.innerWidth || 1365);
defineGetter(window, 'outerHeight', () => (window.innerHeight || 768) + 85);
const originalQuery = window.navigator.permissions && window.navigator.permissions.query;
if (originalQuery) {
	window.navigator.permissions.query = (parameters) => (
		parameters && parameters.name === 'notifications'
			? Promise.resolve({ state: Notification.permission })
			: originalQuery.call(window.navigator.permissions, parameters)
	);
}
const getParameter = WebGLRenderingContext.prototype.getParameter;
WebGLRenderingContext.prototype.getParameter = function(parameter) {
	if (parameter === 37445) return 'Google Inc. (Intel)';
	if (parameter === 37446) return 'ANGLE (Intel, Intel(R) UHD Graphics Direct3D11 vs_5_0 ps_5_0, D3D11)';
	return getParameter.call(this, parameter);
};
if (window.WebGL2RenderingContext) {
	const getParameter2 = WebGL2RenderingContext.prototype.getParameter;
	WebGL2RenderingContext.prototype.getParameter = function(parameter) {
		if (parameter === 37445) return 'Google Inc. (Intel)';
		if (parameter === 37446) return 'ANGLE (Intel, Intel(R) UHD Graphics Direct3D11 vs_5_0 ps_5_0, D3D11)';
		return getParameter2.call(this, parameter);
	};
}
const originalToString = Function.prototype.toString;
Function.prototype.toString = function() {
	if (this === window.navigator.permissions.query) {
		return 'function query() { [native code] }';
	}
	return originalToString.call(this);
};
`)
}

func acceptLanguageForLocale(locale string) string {
	if locale == "" {
		return "en-US,en;q=0.9"
	}
	base := locale
	if i := strings.IndexByte(base, '-'); i > 0 {
		base = base[:i]
	}
	return locale + "," + base + ";q=0.9,en;q=0.8"
}

func quoteJS(s string) string {
	return strconv.Quote(s)
}

func extractChromeVersion(ua string) (major, full string) {
	matches := chromeVersionRe.FindStringSubmatch(ua)
	if len(matches) < 3 {
		return "", ""
	}
	return matches[1], matches[1] + "." + matches[2]
}

func extractEdgeVersion(ua string) (major, full string) {
	matches := edgeVersionRe.FindStringSubmatch(ua)
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

func detectPlatformVersion() string {
	switch runtime.GOOS {
	case "windows":
		return "15.0.0"
	case "darwin":
		return "15.0.0"
	default:
		return "6.6.0"
	}
}
