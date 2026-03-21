package bonk

type textMatchConfig struct {
	exact bool
}

// TextMatchOption configures text matching behavior for GetBy* methods.
type TextMatchOption func(*textMatchConfig)

// Exact requires an exact text match instead of substring containment.
func Exact() TextMatchOption {
	return func(c *textMatchConfig) { c.exact = true }
}

func defaultTextMatchConfig() *textMatchConfig {
	return &textMatchConfig{}
}

type getByRoleConfig struct {
	name    string
	hasName bool
	exact   bool
}

// GetByRoleOption configures GetByRole behavior.
type GetByRoleOption func(*getByRoleConfig)

// WithName filters elements by their accessible name.
// Pass Exact() to require an exact name match.
func WithName(name string, opts ...TextMatchOption) GetByRoleOption {
	return func(c *getByRoleConfig) {
		c.name = name
		c.hasName = true
		tc := defaultTextMatchConfig()
		for _, o := range opts {
			o(tc)
		}
		c.exact = tc.exact
	}
}

func defaultGetByRoleConfig() *getByRoleConfig {
	return &getByRoleConfig{}
}
