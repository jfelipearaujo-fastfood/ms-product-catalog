package environment

import (
	"context"
)

type ApiConfig struct {
	Port       int    `env:"PORT, default=5000"`
	EnvName    string `env:"ENV_NAME, default=development"`
	ApiVersion string `env:"VERSION, default=v1"`
}

func (c *ApiConfig) IsDevelopment() bool {
	return c.EnvName == "development"
}

type DatabaseConfig struct {
	DbName        string `env:"NAME, required"`
	Url           string `env:"URL, required"`
	UrlSecretName string `env:"URL_SECRET_NAME, required"`
}

type CloudConfig struct {
	BaseEndpoint string `env:"BASE_ENDPOINT"`
}

func (c *CloudConfig) IsBaseEndpointSet() bool {
	return c.BaseEndpoint != ""
}

type Config struct {
	ApiConfig   *ApiConfig      `env:",prefix=API_"`
	DbConfig    *DatabaseConfig `env:",prefix=DB_"`
	CloudConfig *CloudConfig    `env:",prefix=AWS_"`
}

type Environment interface {
	GetEnvironmentFromFile(ctx context.Context, fileName string) (*Config, error)
	GetEnvironment(ctx context.Context) (*Config, error)
}
