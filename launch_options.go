package bonk

// LaunchOption configures browser launch behavior.
type LaunchOption func(*launchConfig)

type launchConfig struct {
	headless    bool
	stealth     bool
	chromePath  string
	userDataDir string
	extraArgs   []string
	extraEnv    []string
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
