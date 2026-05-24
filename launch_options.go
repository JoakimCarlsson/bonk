package bonk

import (
	"io"
	"log/slog"
)

// LaunchOption configures browser launch behavior.
type LaunchOption func(*launchConfig)

type launchConfig struct {
	headless    bool
	stealth     bool
	chromePath  string
	userDataDir string
	extraArgs   []string
	extraEnv    []string
	logger      *slog.Logger
	stderrSink  io.Writer
	onCrashpad  func(reports []CrashReport)
}

func defaultLaunchConfig() *launchConfig {
	return &launchConfig{
		headless: true,
		stealth:  true,
	}
}

// Stealth enables anti-detection measures. Default is true.
// When enabled: disables automation-revealing Chrome flags, skips Runtime.enable
// (which is the primary CDP detection signal), injects JS patches to spoof
// navigator properties, and ensures Client Hints headers are consistent.
// Trade-off: OnConsole event handler won't work in stealth mode.
func Stealth(v bool) LaunchOption {
	return func(c *launchConfig) {
		c.stealth = v
	}
}

// Headless controls whether Chrome runs in headless mode. Default is true.
func Headless(v bool) LaunchOption {
	return func(c *launchConfig) {
		c.headless = v
	}
}

// ChromePath sets an explicit path to the Chrome binary.
func ChromePath(path string) LaunchOption {
	return func(c *launchConfig) {
		c.chromePath = path
	}
}

// UserDataDir sets the user data directory for Chrome.
// If empty, a temporary directory is created and cleaned up on Close.
func UserDataDir(path string) LaunchOption {
	return func(c *launchConfig) {
		c.userDataDir = path
	}
}

// Args appends additional command-line arguments to the Chrome launch.
func Args(args ...string) LaunchOption {
	return func(c *launchConfig) {
		c.extraArgs = append(c.extraArgs, args...)
	}
}

// Env appends additional environment variables for the Chrome process.
func Env(env ...string) LaunchOption {
	return func(c *launchConfig) {
		c.extraEnv = append(c.extraEnv, env...)
	}
}

// WithLogger enables debug logging of all CDP messages using the given logger.
func WithLogger(l *slog.Logger) LaunchOption {
	return func(c *launchConfig) {
		c.logger = l
	}
}

// WithStderrSink streams Chrome's stderr to w after the DevTools URL has been
// parsed. Without this option bonk drops the rest of Chrome's stderr on the
// floor, which hides renderer crashes, GPU process errors, and Chrome's own
// "[ERROR]" lines. Pass os.Stderr or a *os.File to a log path.
func WithStderrSink(w io.Writer) LaunchOption {
	return func(c *launchConfig) {
		c.stderrSink = w
	}
}

// WithCrashpadHandler registers a callback fired on Browser.Close with any
// crash dumps that appeared in the user-data-dir's Crashpad/reports directory
// during this session. Useful for surfacing the file path + a one-line summary
// (e.g. v8-oom-location) instead of letting renderer crashes vanish silently.
func WithCrashpadHandler(fn func(reports []CrashReport)) LaunchOption {
	return func(c *launchConfig) {
		c.onCrashpad = fn
	}
}
