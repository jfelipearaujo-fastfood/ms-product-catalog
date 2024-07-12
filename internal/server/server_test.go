package server

import (
	"testing"

	"github.com/jfelipearaujo-org/ms-product-catalog/internal/environment"
	"github.com/stretchr/testify/assert"
)

func TestNewServer(t *testing.T) {
	t.Run("Should return a new server", func(t *testing.T) {
		// Arrange
		config := &environment.Config{
			ApiConfig: &environment.ApiConfig{
				Port: 5000,
			},
			DbConfig: &environment.DatabaseConfig{
				DbName:        "db",
				Url:           "mongodb://host:1234",
				UrlSecretName: "db-secret-url",
			},
		}

		// Act
		server := NewServer(config)

		// Assert
		assert.NotNil(t, server)
		assert.Equal(t, ":5000", server.Addr)
	})
}
