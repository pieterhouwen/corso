package storage

import (
	"github.com/pkg/errors"

	"github.com/alcionai/corso/pkg/credentials"
)

type CommonConfig struct {
	credentials.Corso // requires: CorsoPassword
}

// config key consts
const (
	keyCommonCorsoPassword = "common_corsoPassword"
)

func (c CommonConfig) Config() (config, error) {
	cfg := config{
		keyCommonCorsoPassword: c.CorsoPassword,
	}
	return cfg, c.validate()
}

// CommonConfig retrieves the CommonConfig details from the Storage config.
func (s Storage) CommonConfig() (CommonConfig, error) {
	c := CommonConfig{}
	if len(s.Config) > 0 {
		c.CorsoPassword = orEmptyString(s.Config[keyCommonCorsoPassword])
	}
	return c, c.validate()
}

// ensures all required properties are present
func (c CommonConfig) validate() error {
	if len(c.CorsoPassword) == 0 {
		return errors.Wrap(errMissingRequired, credentials.CORSO_PASSWORD)
	}
	return nil
}
