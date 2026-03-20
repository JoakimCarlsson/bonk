package bonk

// LaunchOption configures browser launch behavior.
type LaunchOption func(*launchConfig)

type launchConfig struct {
	headless    bool
	chromePath  string
	userDataDir string
	extraArgs   []string
	extraEnv    []string
}

func defaultLaunchConfig() *launchConfig {
	return &launchConfig{
		headless: true,
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
