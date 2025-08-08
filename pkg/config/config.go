package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Data             interface{}
	DefaultFilePaths []string
	DefaultFileNames []string
}

func (cnf *Config) IsTestFlag() bool {
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "-test.") {
			return true
		}
	}

	return false
}

func (cnf *Config) Validate() error {
	validate := validator.New()
	validateError := validate.Struct(cnf.Data)

	if validateError != nil {
		return validateError
	}

	return nil
}

func (cnf *Config) ReadFromEnv() error {
	err := cleanenv.ReadEnv(cnf.Data)
	if err != nil {
		return err
	}

	return nil
}

func (cnf *Config) ReadFromFile() {
	for _, file := range cnf.getConfigFile() {
		absPath, _ := filepath.Abs(file)
		err := cleanenv.ReadConfig(absPath, cnf.Data)
		if err != nil {
			err2 := errors.Unwrap(err)
			if err2 == nil || err2.Error() != "no such file or directory" {
				fmt.Printf("Reading config was failed from: %v with err: %v\n", absPath, err)
			}
		} else {
			fmt.Printf("Reading config was successful from: %v\n", absPath)
		}
	}
}

func (cnf *Config) getConfigFile() []string {
	files := make([]string, 0)
	for _, path := range cnf.getConfigPaths() {
		for _, fileName := range cnf.getConfigFileNames() {
			files = append(files, fmt.Sprintf("%s/%s", path, fileName))
		}
	}

	return files
}

func (cnf *Config) getConfigPaths() []string {
	if envPath := os.Getenv("CONFIG_FILE_PATH"); envPath != "" {
		return []string{envPath}
	}
	return cnf.DefaultFilePaths
}

func (cnf *Config) getConfigFileNames() []string {
	if envName := os.Getenv("CONFIG_FILE_NAME"); envName != "" {
		return []string{envName}
	}
	return cnf.DefaultFileNames
}
