package bootstrap

import (
	"fmt"

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
		DefaultFileNames: []string{"config.json"},
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

	fmt.Println("==========")
	fmt.Println(configInt.HTTPServer.BindAddress)
	fmt.Println(configInt.DataBase.DatabaseDsn)
	fmt.Println(configInt.CalculateSystem.URI)
	fmt.Println("==========")

	return configInt, nil
}
