package config

import "errors"

type queueSevice struct {
	Addr          string `yaml:"addr"`
	Limit         int    `yaml:"limit"`
	TokenLifetime int    `yaml:"tokenLifetime"`
}

func (c *queueSevice) validate() error {
	if c.Limit < 0 {
		return errors.New("invalid limit")
	}

	if c.TokenLifetime < 1 {
		return errors.New("invalid tokenLifetime")
	}
	return nil
}
