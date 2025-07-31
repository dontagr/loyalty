package bootstrap

import (
	"go.uber.org/fx"

	configInternal "github.com/dontagr/loyalty/internal/config"
	"github.com/dontagr/loyalty/pkg/config"
)

var Config = fx.Options(
	fx.Provide(newConfig),
)

func newConfig() (*configInternal.Config, error) {
	configInt := &configInternal.Config{}
	flagEnricher := &configInternal.FlagEnricher{}

	cnf := &config.Config{
		Data:             configInt,
		DefaultFilePaths: []string{"../../../configs", "./configs"},
		DefaultFileNames: []string{"config.json", "config.dev.json"},
	}

	cnf.ReadFromFile()
	if !cnf.IsTestFlag() {
		err := flagEnricher.Process(configInt)
		if err != nil {
			return nil, err
		}
	}

	err := cnf.ReadFromEnv()
	if err != nil {
		return nil, err
	}

	err = cnf.Validate()
	if err != nil {
		return nil, err
	}

	return configInt, nil
}
