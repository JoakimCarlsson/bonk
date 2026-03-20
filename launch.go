package bonk

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
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

	conn := rpc.New(ws)
	go conn.Listen(ctx)

	return &Browser{
		conn:    conn,
		process: cmd.Process,
		cmd:     cmd,
		dataDir: dataDir,
		tempDir: tempDir,
		wsURL:   wsURL,
		ctx:     ctx,
		cancel:  cancel,
	}, nil
}

// Connect attaches to an already-running Chrome instance.
func Connect(wsURL string) (*Browser, error) {
	ctx, cancel := context.WithCancel(context.Background())

	ws, err := transport.Dial(ctx, wsURL)
	if err != nil {
		cancel()
		return nil, err
	}

	conn := rpc.New(ws)
	go conn.Listen(ctx)

	return &Browser{
		conn:   conn,
		wsURL:  wsURL,
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

func buildArgs(cfg *launchConfig, dataDir string) []string {
	args := make([]string, len(defaultChromeFlags))
	copy(args, defaultChromeFlags)

	args = append(args, "--user-data-dir="+dataDir)

	if cfg.headless {
		args = append(args, "--headless=new")
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

func cleanupDir(dir string) {
	if dir != "" {
		os.RemoveAll(dir)
	}
}
