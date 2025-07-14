package utils

import (
	"os"
	"testing"
)

func TestGetEnvReturnsValue(t *testing.T) {
	os.Setenv("FOO", "bar")
	defer os.Unsetenv("FOO")

	got := GetEnv("FOO")
	if got != "bar" {
		t.Errorf("Expected 'bar', got '%s'", got)
	}
}

func TestGetEnvReturnsEmptyIfNotSet(t *testing.T) {
	got := GetEnv("DOES_NOT_EXIST")
	if got != "" {
		t.Errorf("Expected empty string, got '%s'", got)
	}
}
