package environment

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsDevelopment(t *testing.T) {
	t.Run("Should return true if development", func(t *testing.T) {
		// Arrange
		c := &ApiConfig{
			EnvName: "development",
		}

		// Act
		res := c.IsDevelopment()

		// Assert
		assert.True(t, res)
	})

	t.Run("Should return false if not development", func(t *testing.T) {
		// Arrange
		c := &ApiConfig{
			EnvName: "production",
		}

		// Act
		res := c.IsDevelopment()

		// Assert
		assert.False(t, res)
	})
}

func TestIsBaseEndpointSet(t *testing.T) {
	t.Run("Should return true if base endpoint set", func(t *testing.T) {
		// Arrange
		c := &CloudConfig{
			BaseEndpoint: "http://localhost:5000",
		}

		// Act
		res := c.IsBaseEndpointSet()

		// Assert
		assert.True(t, res)
	})

	t.Run("Should return false if base endpoint not set", func(t *testing.T) {
		// Arrange
		c := &CloudConfig{}

		// Act
		res := c.IsBaseEndpointSet()

		// Assert
		assert.False(t, res)
	})
}
