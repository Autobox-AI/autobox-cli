package config

import (
	"os"
	"testing"

	"github.com/spf13/viper"
)

func TestSetDefaults(t *testing.T) {
	viper.Reset()
	setDefaults()

	if viper.GetString("docker.host") != "unix:///var/run/docker.sock" {
		t.Errorf("docker.host: got %s, want unix:///var/run/docker.sock", viper.GetString("docker.host"))
	}

	if viper.GetString("docker.image") != "autobox-engine:latest" {
		t.Errorf("docker.image: got %s, want autobox-engine:latest", viper.GetString("docker.image"))
	}

	if viper.GetString("output.format") != "table" {
		t.Errorf("output.format: got %s, want table", viper.GetString("output.format"))
	}

	if viper.GetBool("output.color") != true {
		t.Errorf("output.color: got %v, want true", viper.GetBool("output.color"))
	}
}

func TestInit(t *testing.T) {
	os.Setenv("AUTOBOX_DOCKER_HOST", "tcp://localhost:2375")
	defer os.Unsetenv("AUTOBOX_DOCKER_HOST")

	viper.Reset()
	err := Init()
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	if viper.GetString("docker.host") != "tcp://localhost:2375" {
		t.Errorf("docker.host: got %s, want tcp://localhost:2375", viper.GetString("docker.host"))
	}
}

func TestGetString(t *testing.T) {
	viper.Reset()
	viper.Set("test.key", "test-value")

	result := GetString("test.key")
	if result != "test-value" {
		t.Errorf("GetString: got %s, want test-value", result)
	}
}

func TestGetBool(t *testing.T) {
	viper.Reset()
	viper.Set("test.bool", true)

	result := GetBool("test.bool")
	if result != true {
		t.Errorf("GetBool: got %v, want true", result)
	}
}

func TestGetInt(t *testing.T) {
	viper.Reset()
	viper.Set("test.int", 42)

	result := GetInt("test.int")
	if result != 42 {
		t.Errorf("GetInt: got %d, want 42", result)
	}
}

func TestGetStringSlice(t *testing.T) {
	viper.Reset()
	expected := []string{"item1", "item2", "item3"}
	viper.Set("test.slice", expected)

	result := GetStringSlice("test.slice")
	if len(result) != len(expected) {
		t.Errorf("GetStringSlice: got length %d, want %d", len(result), len(expected))
	}

	for i, v := range result {
		if v != expected[i] {
			t.Errorf("GetStringSlice[%d]: got %s, want %s", i, v, expected[i])
		}
	}
}

func TestGetStringMap(t *testing.T) {
	viper.Reset()
	expected := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	}
	viper.Set("test.map", expected)

	result := GetStringMap("test.map")
	if len(result) != len(expected) {
		t.Errorf("GetStringMap: got length %d, want %d", len(result), len(expected))
	}

	if result["key1"] != "value1" {
		t.Errorf("GetStringMap[key1]: got %s, want value1", result["key1"])
	}
}
