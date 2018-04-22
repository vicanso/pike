package config

import (
	"testing"
)

func TestConfig(t *testing.T) {
	t.Run("get config", func(t *testing.T) {
		config := GetDefault()
		if config.Name != "Pike" {
			t.Fatalf("init name fail")
		}
		if len(config.Directors) != 2 {
			t.Fatalf("init directors fail")
		}
	})
}
