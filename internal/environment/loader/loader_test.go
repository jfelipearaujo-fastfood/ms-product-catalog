package loader

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/jfelipearaujo-org/ms-product-catalog/internal/environment"
	"github.com/stretchr/testify/assert"
)

func cleanEnv() {
	os.Unsetenv("API_PORT")
	os.Unsetenv("API_ENV_NAME")
	os.Unsetenv("API_VERSION")

	os.Unsetenv("DB_NAME")
	os.Unsetenv("DB_URL")
	os.Unsetenv("DB_URL_SECRET_NAME")
}

func TestGetEnvironment(t *testing.T) {
	t.Run("Should load environment variables", func(t *testing.T) {
		// Arrange
		t.Setenv("API_PORT", "5000")
		t.Setenv("API_ENV_NAME", "development")
		t.Setenv("API_VERSION", "v1")

		t.Setenv("DB_NAME", "test")
		t.Setenv("DB_URL", "db://host:1234")
		t.Setenv("DB_URL_SECRET_NAME", "db-secret-url")

		expected := &environment.Config{
			ApiConfig: &environment.ApiConfig{
				Port:       5000,
				EnvName:    "development",
				ApiVersion: "v1",
			},
			DbConfig: &environment.DatabaseConfig{
				DbName:        "test",
				Url:           "db://host:1234",
				UrlSecretName: "db-secret-url",
			},
			CloudConfig: &environment.CloudConfig{
				BaseEndpoint: "",
			},
		}

		// Act
		env, err := NewLoader().GetEnvironment(context.Background())

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, env)
		assert.Equal(t, expected, env)
	})

	t.Run("Should return error if a required variable is not set", func(t *testing.T) {
		// Arrange
		t.Setenv("API_PORT", "5000")
		t.Setenv("API_ENV_NAME", "development")
		t.Setenv("API_VERSION", "v1")

		t.Setenv("DB_NAME", "test")

		// Act
		env, err := NewLoader().GetEnvironment(context.Background())

		// Assert
		assert.Error(t, err)
		assert.Nil(t, env)
	})
}

func TestGetEnvironmentFromFile(t *testing.T) {
	t.Run("Should load environment variables from file", func(t *testing.T) {
		// Arrange
		cleanEnv()

		expected := &environment.Config{
			ApiConfig: &environment.ApiConfig{
				Port:       5000,
				EnvName:    "development",
				ApiVersion: "v1",
			},
			DbConfig: &environment.DatabaseConfig{
				DbName:        "test",
				Url:           "db://host:1234",
				UrlSecretName: "db-secret-url",
			},
			CloudConfig: &environment.CloudConfig{
				BaseEndpoint: "",
			},
		}

		// Act
		env, err := NewLoader().GetEnvironmentFromFile(context.Background(), filepath.Join("testdata", "test.env"))

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, env)
		assert.Equal(t, expected, env)
	})

	t.Run("Should return error if a required variable is not set", func(t *testing.T) {
		// Arrange
		cleanEnv()

		// Act
		env, err := NewLoader().GetEnvironmentFromFile(context.Background(), filepath.Join("testdata", "test_err.env"))

		// Assert
		assert.Error(t, err)
		assert.Nil(t, env)
	})

	t.Run("Should return error try to load from an invalid file", func(t *testing.T) {
		// Arrange
		cleanEnv()

		// Act
		env, err := NewLoader().GetEnvironmentFromFile(context.Background(), filepath.Join("testdata", "non_exists.env"))

		// Assert
		assert.Error(t, err)
		assert.Nil(t, env)
	})
}
