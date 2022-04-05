package domain

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewApplicationName_ValidValue(t *testing.T) {
	appName, err := NewApplicationName("test")
	require.NoError(t, err)
	assert.EqualValues(t, "test", *appName)
}

func TestNewApplicationName_EmptyValue_ReturnsError(t *testing.T) {
	_, err := NewApplicationName("")
	require.Error(t, err)
}

func TestNewURL_ValidValue(t *testing.T) {
	url, err := NewURL("https://google.com")
	require.NoError(t, err)
	assert.EqualValues(t, "https://google.com", *url)
}

func TestNewURL_InvalidURL_ReturnsError(t *testing.T) {
	_, err := NewURL("not-a-url")
	require.Error(t, err)
}

func TestNewURL_EmptyValue_ReturnsError(t *testing.T) {
	_, err := NewURL("")
	require.Error(t, err)
}

func TestNewRating_ValidValue(t *testing.T) {
	rating, err := NewRating(3.89)
	require.NoError(t, err)
	assert.EqualValues(t, 3.89, *rating)
}

func TestNewRating_NegativeValue_ReturnsError(t *testing.T) {
	_, err := NewRating(-1)
	require.Error(t, err)
}
