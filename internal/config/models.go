package config

type Config struct {
	Log             Logging         `json:"Logging"`
	HTTPServer      HTTPServer      `json:"HttpServing"`
	DataBase        DataBase        `json:"DataBase"`
	Security        Security        `json:"Security"`
	CalculateSystem CalculateSystem `json:"CalculateSystem"`
}

type CalculateSystem struct {
	URI string `json:"URI" env:"ACCRUAL_SYSTEM_ADDRESS" flag:"r" validate:"required"`
}

type DataBase struct {
	DatabaseDsn string `json:"DatabaseDsn" env:"DATABASE_DSN" flag:"d" validate:"required"`
}

type HTTPServer struct {
	BindAddress string `json:"BindAddress" env:"RUN_ADDRESS" flag:"a" validate:"required"`
}

type Logging struct {
	LogLevel string `json:"LogLevel" validate:"required"`
}

type Security struct {
	Key string `json:"key" validate:"required"`
}
