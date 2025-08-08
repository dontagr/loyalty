package config

import (
	"flag"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type FlagEnricher struct {
}

func (f *FlagEnricher) Process(cnf *Config) error {
	flagSet := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	flagSet.SetOutput(os.Stderr)
	flagSet.Usage = cleanenv.FUsage(flagSet.Output(), cnf, nil, flagSet.Usage)

	serverAddrBind := flagSet.String("a", "", "bind addr http")
	databaseDsn := flagSet.String("d", "", "string with the database connection address")
	calculateSystemURI := flagSet.String("r", "", "address of the accrual calculation system")

	err := flagSet.Parse(os.Args[1:])
	if err != nil {
		return err
	}
	if *serverAddrBind != "" {
		cnf.HTTPServer.BindAddress = *serverAddrBind
	}
	if *databaseDsn != "" {
		cnf.DataBase.DatabaseDsn = *databaseDsn
	}
	if *calculateSystemURI != "" {
		cnf.CalculateSystem.URI = *calculateSystemURI
	}

	return nil
}
