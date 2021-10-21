package config

type Config interface {
	Setting() Setting
}

type DefaultConfig struct {
	setting Setting
}

var _ Config = (*DefaultConfig)(nil)

func (c *DefaultConfig) Setting() Setting {
	return c.setting
}

func New(setting Setting) *DefaultConfig {
	return &DefaultConfig{
		setting: setting,
	}
}
