package config

import (
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	DefaultFilePaths = []string{"../../../configs", "../../configs", "./configs"}
	DefaultFileNames = []string{"config.test.json"}
)

type ConfigTest struct {
	Service Service `json:"service"`
}
type Service struct {
	WorkerLimit int `json:"WorkerLimit"`
}

type EnvConfig struct {
	Port     string `env:"PORT"`
	Database string `env:"DATABASE"`
}

func TestConfig_IsTestFlag(t *testing.T) {
	type fields struct {
		Data             interface{}
		DefaultFilePaths []string
		DefaultFileNames []string
	}
	tests := []struct {
		name   string
		fields fields
		args   []string
		want   bool
	}{
		{
			name: "Test flag present",
			fields: fields{
				Data:             nil,
				DefaultFilePaths: []string{},
				DefaultFileNames: []string{},
			},
			args: []string{"app", "-test.v", "-otherFlag"},
			want: true,
		},
		{
			name: "Test flag absent",
			fields: fields{
				Data:             nil,
				DefaultFilePaths: []string{},
				DefaultFileNames: []string{},
			},
			args: []string{"app", "-otherFlag"},
			want: false,
		},
		{
			name: "Multiple test flags",
			fields: fields{
				Data:             nil,
				DefaultFilePaths: []string{},
				DefaultFileNames: []string{},
			},
			args: []string{"app", "-test.run", "-test.cover"},
			want: true,
		},
		{
			name: "Empty args",
			fields: fields{
				Data:             nil,
				DefaultFilePaths: []string{},
				DefaultFileNames: []string{},
			},
			args: []string{},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = tt.args
			cnf := &Config{
				Data:             tt.fields.Data,
				DefaultFilePaths: tt.fields.DefaultFilePaths,
				DefaultFileNames: tt.fields.DefaultFileNames,
			}
			if got := cnf.IsTestFlag(); got != tt.want {
				t.Errorf("IsTestFlag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_ReadFromEnv(t *testing.T) {
	type fields struct {
		Data             interface{}
		DefaultFilePaths []string
		DefaultFileNames []string
	}
	tests := []struct {
		name    string
		fields  fields
		envVars map[string]string
		wantErr bool
		want    EnvConfig
	}{
		{
			name: "Переменные окружения установлены",
			fields: fields{
				Data:             &EnvConfig{},
				DefaultFilePaths: []string{},
				DefaultFileNames: []string{},
			},
			envVars: map[string]string{
				"PORT":     "8080",
				"DATABASE": "postgres",
			},
			wantErr: false,
			want:    EnvConfig{Port: "8080", Database: "postgres"},
		},
		{
			name: "Переменные окружения не установлены",
			fields: fields{
				Data:             &EnvConfig{},
				DefaultFilePaths: []string{},
				DefaultFileNames: []string{},
			},
			envVars: map[string]string{}, // Переменные окружения не установлены
			wantErr: false,
			want:    EnvConfig{}, // Ожидаются стандартные значения, так как переменные окружения не заданы
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cnf := &Config{
				Data:             tt.fields.Data,
				DefaultFilePaths: tt.fields.DefaultFilePaths,
				DefaultFileNames: tt.fields.DefaultFileNames,
			}
			if err := cnf.ReadFromEnv(); (err != nil) != tt.wantErr {
				t.Errorf("ReadFromEnv() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_ReadFromFile(t *testing.T) {
	t.Run("valid configuration", func(t *testing.T) {
		cfg := &Config{
			Data:             &ConfigTest{},
			DefaultFilePaths: DefaultFilePaths,
			DefaultFileNames: DefaultFileNames,
		}
		cfg.ReadFromFile()

		c, ok := cfg.Data.(*ConfigTest)
		if !ok {
			t.Errorf("config not valid")
			return
		}

		assert.Equal(t, 999, c.Service.WorkerLimit)
	})
}

func TestConfig_Validate(t *testing.T) {
	type fields struct {
		Data             interface{}
		DefaultFilePaths []string
		DefaultFileNames []string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "valid config",
			fields: fields{
				Data: &ConfigTest{
					Service: Service{
						WorkerLimit: 999,
					},
				},
				DefaultFilePaths: DefaultFilePaths,
				DefaultFileNames: DefaultFileNames,
			},
			wantErr: false,
		},
		{
			name: "empty data",
			fields: fields{
				Data:             nil,
				DefaultFilePaths: DefaultFilePaths,
				DefaultFileNames: DefaultFileNames,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cnf := &Config{
				Data:             tt.fields.Data,
				DefaultFilePaths: tt.fields.DefaultFilePaths,
				DefaultFileNames: tt.fields.DefaultFileNames,
			}
			if err := cnf.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_getConfigFile(t *testing.T) {
	type fields struct {
		Data             interface{}
		DefaultFilePaths []string
		DefaultFileNames []string
		EnvSet           bool
		EnvFilePaths     string
		EnvFileNames     string
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{
			name: "existing config file",
			fields: fields{
				DefaultFilePaths: DefaultFilePaths,
				DefaultFileNames: DefaultFileNames,
				EnvSet:           false,
			},
			want: []string{"../../../configs/config.test.json", "../../configs/config.test.json", "./configs/config.test.json"},
		},
		{
			name: "existing config file",
			fields: fields{
				DefaultFilePaths: []string{},
				DefaultFileNames: []string{},
				EnvSet:           true,
				EnvFilePaths:     "test/test",
				EnvFileNames:     "conf.json",
			},
			want: []string{"test/test/conf.json"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.fields.EnvSet {
				err := os.Setenv("CONFIG_FILE_PATH", tt.fields.EnvFilePaths)
				if err != nil {
					t.Errorf("failed set env CONFIG_FILE_PATH %v", err)
					return
				}
				err = os.Setenv("CONFIG_FILE_NAME", tt.fields.EnvFileNames)
				if err != nil {
					t.Errorf("failed set env CONFIG_FILE_NAME %v", err)
					return
				}
			}

			cnf := &Config{
				Data:             tt.fields.Data,
				DefaultFilePaths: tt.fields.DefaultFilePaths,
				DefaultFileNames: tt.fields.DefaultFileNames,
			}
			if got := cnf.getConfigFile(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getConfigFile() = %v, want %v", got, tt.want)
			}
		})
	}
}
