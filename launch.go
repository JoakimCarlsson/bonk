package bonk

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/joakimcarlsson/bonk/rpc"
	"github.com/joakimcarlsson/bonk/transport"
)

var defaultChromeFlags = []string{
	"--no-first-run",
	"--no-default-browser-check",
	"--disable-background-networking",
	"--disable-background-timer-throttling",
	"--disable-backgrounding-occluded-windows",
	"--disable-breakpad",
	"--disable-component-extensions-with-background-pages",
	"--disable-default-apps",
	"--disable-dev-shm-usage",
	"--disable-extensions",
	"--disable-hang-monitor",
	"--disable-ipc-flooding-protection",
	"--disable-popup-blocking",
	"--disable-prompt-on-repost",
	"--disable-renderer-backgrounding",
	"--disable-sync",
	"--disable-translate",
	"--metrics-recording-only",
	"--no-startup-window",
	"--password-store=basic",
	"--use-mock-keychain",
	"--remote-debugging-port=0",
}

// Launch starts a new Chrome browser process and connects to it.
func Launch(opts ...LaunchOption) (*Browser, error) {
	cfg := defaultLaunchConfig()
	for _, o := range opts {
		o(cfg)
	}

	chromePath, err := findChrome(cfg.chromePath)
	if err != nil {
		return nil, err
	}

	cfg.chromium = isChromium(chromePath)
	if cfg.stealth && cfg.headless && !hasArg(cfg.extraArgs, "--user-agent") {
		if ua := headfulUserAgent(chromePath); ua != "" {
			cfg.userAgent = ua
		}
	}

	tempDir := ""
	dataDir := cfg.userDataDir
	if dataDir == "" {
		tempDir, err = os.MkdirTemp("", "bonk-chrome-*")
		if err != nil {
			return nil, fmt.Errorf("bonk: create temp dir: %w", err)
		}
		dataDir = tempDir
	}

	args := buildArgs(cfg, dataDir)

	cmd := exec.Command(chromePath, args...)
	cmd.Env = append(os.Environ(), cfg.extraEnv...)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		cleanupDir(tempDir)
		return nil, fmt.Errorf("bonk: stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		cleanupDir(tempDir)
		return nil, fmt.Errorf("bonk: start chrome: %w", err)
	}

	wsURL, err := parseDevToolsURL(stderr, 30*time.Second)
	if err != nil {
		cmd.Process.Kill()
		cmd.Wait()
		cleanupDir(tempDir)
		return nil, fmt.Errorf("%w: %v", ErrChromeStartup, err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	ws, err := transport.Dial(ctx, wsURL)
	if err != nil {
		cancel()
		cmd.Process.Kill()
		cmd.Wait()
		cleanupDir(tempDir)
		return nil, err
	}

	var t transport.Transport = ws
	if cfg.logger != nil {
		t = &transport.Debug{Inner: ws, Logger: cfg.logger}
	}

	conn := rpc.New(t)
	go conn.Listen(ctx)

	b := &Browser{
		conn:    conn,
		process: cmd.Process,
		cmd:     cmd,
		dataDir: dataDir,
		tempDir: tempDir,
		wsURL:   wsURL,
		stealth: cfg.stealth,
		cfg:     cfg,
		ctx:     ctx,
		cancel:  cancel,
	}
	b.setupOnClose(conn)
	return b, nil
}

// Connect attaches to an already-running Chrome instance.
func Connect(wsURL string, opts ...LaunchOption) (*Browser, error) {
	cfg := defaultLaunchConfig()
	for _, o := range opts {
		o(cfg)
	}

	ctx, cancel := context.WithCancel(context.Background())

	ws, err := transport.Dial(ctx, wsURL)
	if err != nil {
		cancel()
		return nil, err
	}

	var t transport.Transport = ws
	if cfg.logger != nil {
		t = &transport.Debug{Inner: ws, Logger: cfg.logger}
	}

	conn := rpc.New(t)
	go conn.Listen(ctx)

	b := &Browser{
		conn:   conn,
		wsURL:  wsURL,
		cfg:    cfg,
		ctx:    ctx,
		cancel: cancel,
	}
	b.setupOnClose(conn)
	return b, nil
}

// stealthDroppedFlags are default flags that page scripts can observe, so a
// stealth launch leaves them out: without popup blocking a script can open a
// window unprompted, and without IPC flooding protection history.replaceState
// is never throttled.
var stealthDroppedFlags = map[string]bool{
	"--disable-popup-blocking":          true,
	"--disable-ipc-flooding-protection": true,
}

// buildArgs returns the command line flags Chrome is started with for cfg,
// using dataDir as its profile directory.
func buildArgs(cfg *launchConfig, dataDir string) []string {
	args := make([]string, 0, len(defaultChromeFlags))
	for _, f := range defaultChromeFlags {
		if cfg.stealth && stealthDroppedFlags[f] {
			continue
		}
		args = append(args, f)
	}

	args = append(args, "--user-data-dir="+dataDir)

	if cfg.headless {
		args = append(args, "--headless=new")
	}

	if cfg.stealth {
		args = append(args,
			"--disable-blink-features=AutomationControlled",
			"--disable-features=IsolateOrigins,site-per-process",
			"--disable-infobars",
			"--disable-session-crashed-bubble",
			"--disable-search-engine-choice-screen",
		)
		if cfg.headless && cfg.userAgent != "" {
			args = append(args, "--user-agent="+cfg.userAgent)
		}
		if cfg.headless {
			args = append(args,
				"--window-size=1680,960",
				"--screen-info={1920x1080}",
				"--enable-gpu",
				"--blink-settings=primaryHoverType=2,availableHoverTypes=2,"+
					"primaryPointerType=4,availablePointerTypes=4",
			)
		}
	}

	args = append(args, cfg.extraArgs...)
	return args
}

func parseDevToolsURL(
	r interface{ Read([]byte) (int, error) },
	timeout time.Duration,
) (string, error) {
	scanner := bufio.NewScanner(r)
	done := make(chan string, 1)
	errCh := make(chan error, 1)

	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "DevTools listening on ") {
				url := strings.TrimPrefix(line, "DevTools listening on ")
				done <- strings.TrimSpace(url)
				return
			}
		}
		if err := scanner.Err(); err != nil {
			errCh <- err
		} else {
			errCh <- fmt.Errorf("chrome closed without providing DevTools URL")
		}
	}()

	select {
	case url := <-done:
		return url, nil
	case err := <-errCh:
		return "", err
	case <-time.After(timeout):
		return "", fmt.Errorf("timeout waiting for DevTools URL")
	}
}

func findChrome(explicit string) (string, error) {
	if explicit != "" {
		if _, err := os.Stat(explicit); err == nil {
			return explicit, nil
		}
		return "", fmt.Errorf("%w: %s not found", ErrChromeNotFound, explicit)
	}

	if p := os.Getenv("CHROME_PATH"); p != "" {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}

	candidates := chromeCandidates()
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c, nil
		}
	}

	for _, name := range []string{"google-chrome-stable", "google-chrome", "chromium-browser", "chromium"} {
		if p, err := exec.LookPath(name); err == nil {
			return p, nil
		}
	}

	return "", ErrChromeNotFound
}

func chromeCandidates() []string {
	switch runtime.GOOS {
	case "linux":
		return []string{
			"/usr/bin/google-chrome-stable",
			"/usr/bin/google-chrome",
			"/usr/bin/chromium-browser",
			"/usr/bin/chromium",
			"/snap/bin/chromium",
		}
	case "darwin":
		return []string{
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Chromium.app/Contents/MacOS/Chromium",
		}
	case "windows":
		localAppData := os.Getenv("LOCALAPPDATA")
		programFiles := os.Getenv("PROGRAMFILES")
		return []string{
			programFiles + `\Google\Chrome\Application\chrome.exe`,
			localAppData + `\Google\Chrome\Application\chrome.exe`,
		}
	}
	return nil
}

// versionRe finds the version number in the output of chrome --version.
var versionRe = regexp.MustCompile(`(\d+)\.\d+\.\d+\.\d+`)

// headfulUserAgent returns the user agent a headful Chrome of the binary at
// path sends, for headless Chrome to send from every page and worker; the
// page-level override stealth applies does not reach service workers. The
// version is read from the binary's --version output, and the platform part
// is the reduced one Chrome sends on the host OS. It returns an empty
// string when the version cannot be read.
func headfulUserAgent(path string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, path, "--version").Output()
	if err != nil {
		return ""
	}
	m := versionRe.FindStringSubmatch(string(out))
	if m == nil {
		return ""
	}
	return "Mozilla/5.0 (" + reducedPlatform() + ") AppleWebKit/537.36 " +
		"(KHTML, like Gecko) Chrome/" + m[1] + ".0.0.0 Safari/537.36"
}

// reducedPlatform returns the platform part of the reduced user agent
// Chrome sends on the host OS.
func reducedPlatform() string {
	switch runtime.GOOS {
	case "darwin":
		return "Macintosh; Intel Mac OS X 10_15_7"
	case "windows":
		return "Windows NT 10.0; Win64; x64"
	case "android":
		return "Linux; Android 10; K"
	default:
		return "X11; Linux x86_64"
	}
}

// hasArg reports whether args holds the flag name, bare or with a value.
func hasArg(args []string, name string) bool {
	for _, a := range args {
		if a == name || strings.HasPrefix(a, name+"=") {
			return true
		}
	}
	return false
}

// isChromium reports whether the browser binary at path is a plain Chromium
// build rather than Google Chrome, judged by its path.
func isChromium(path string) bool {
	lower := strings.ToLower(path)
	return strings.Contains(lower, "chromium") ||
		strings.Contains(lower, "ms-playwright")
}

// cleanupDir removes dir when it is not empty.
func cleanupDir(dir string) {
	if dir != "" {
		os.RemoveAll(dir)
	}
}
