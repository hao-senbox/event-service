package config

import "os"

type Consul struct {
	Host string `mapstructure:"host" validate:"required"`
	Port string `mapstructure:"port" validate:"required"`
}

type Registry struct {
	Host string `mapstructure:"host" validate:"required"`
}

type AppConfiguration struct {
	Name        string    `mapstructure:"name"`
	Version     string    `mapstructure:"version"`
	Environment string    `mapstructure:"environment"`
	API         APIConfig `mapstructure:"api"`
}

type APIConfig struct {
	Rest RestConfig `mapstructure:"rest"`
}

type RestConfig struct {
	Host    string        `mapstructure:"host"`
	Port    string        `mapstructure:"port"`
	Setting SettingConfig `mapstructure:"setting"`
}

type SettingConfig struct {
	Debug               bool     `mapstructure:"debug"`
	DebugErrorsResponse bool     `mapstructure:"debugErrorsResponse"`
	IgnoreLogUrls       []string `mapstructure:"ignoreLogUrls"`
}

type ZapConfig struct {
	Development bool   `mapstructure:"development"`
	Caller      bool   `mapstructure:"caller"`
	Stacktrace  string `mapstructure:"stacktrace"`
	Cores       struct {
		Console struct {
			Type     string `mapstructure:"type"`
			Level    string `mapstructure:"level"`
			Encoding string `mapstructure:"encoding"`
		} `mapstructure:"console"`
	} `mapstructure:"cores"`
}

type Config struct {
	Port     string
	MongoURI string
	MongoDB  string
	Consul   Consul           `mapstructure:"consul" validate:"required"`
	Registry Registry         `mapstructure:"registry" validate:"required"`
	App      AppConfiguration `mapstructure:"app"`
	Zap      ZapConfig        `mapstructure:"zap"`
}

func LoadConfig() *Config {
	config := &Config{
		Port:     getEnv("PORT", "8007"),
		MongoURI: getEnv("MONGO_URI", "mongodb://localhost:27016"),
		MongoDB:  getEnv("MONGO_DB", "portal"),
		Consul: Consul{
			Host: getEnv("CONSUL_HOST", "localhost"),
			Port: getEnv("CONSUL_PORT", "8500"),
		},
		Registry: Registry{
			Host: getEnv("REGISTRY_HOST", "localhost"),
		},
		App: AppConfiguration{
			API: APIConfig{
				Rest: RestConfig{
					Port: getEnv("PORT", "8007"),
				},
			},
		},
		Zap: ZapConfig{
			Development: true,
			Caller:      true,
			Stacktrace:  "error",
			Cores: struct {
				Console struct {
					Type     string `mapstructure:"type"`
					Level    string `mapstructure:"level"`
					Encoding string `mapstructure:"encoding"`
				} `mapstructure:"console"`
			}{
				Console: struct {
					Type     string `mapstructure:"type"`
					Level    string `mapstructure:"level"`
					Encoding string `mapstructure:"encoding"`
				}{
					Type:     "stream",
					Level:    "debug",
					Encoding: "console",
				},
			},
		},
	}
	return config
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
