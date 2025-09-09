package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	Docker       DockerConfig       `mapstructure:"docker"`
	Simulation   SimulationConfig   `mapstructure:"simulation"`
	Output       OutputConfig       `mapstructure:"output"`
}

type DockerConfig struct {
	Host       string `mapstructure:"host"`
	APIVersion string `mapstructure:"api_version"`
	TLSVerify  bool   `mapstructure:"tls_verify"`
	CertPath   string `mapstructure:"cert_path"`
	Image      string `mapstructure:"image"`
}

type SimulationConfig struct {
	DefaultImage      string            `mapstructure:"default_image"`
	DefaultConfigPath string            `mapstructure:"default_config_path"`
	DefaultMetricsPath string           `mapstructure:"default_metrics_path"`
	DefaultVolumes    []string          `mapstructure:"default_volumes"`
	DefaultEnvironment map[string]string `mapstructure:"default_environment"`
	LogsDirectory     string            `mapstructure:"logs_directory"`
	ConfigsDirectory  string            `mapstructure:"configs_directory"`
}

type OutputConfig struct {
	Format  string `mapstructure:"format"`
	Verbose bool   `mapstructure:"verbose"`
	Color   bool   `mapstructure:"color"`
}

var (
	cfg *Config
)

func Init() error {
	viper.SetConfigName("autobox")
	viper.SetConfigType("yaml")
	
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	
	viper.AddConfigPath(filepath.Join(home, ".autobox"))
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/autobox")

	viper.SetEnvPrefix("AUTOBOX")
	viper.AutomaticEnv()

	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("failed to read config file: %w", err)
		}
	}

	cfg = &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}

func setDefaults() {
	viper.SetDefault("docker.host", "unix:///var/run/docker.sock")
	viper.SetDefault("docker.api_version", "1.41")
	viper.SetDefault("docker.tls_verify", false)
	viper.SetDefault("docker.image", "autobox-engine:latest")

	// Get home directory for default paths
	home, _ := os.UserHomeDir()
	defaultConfigsDir := filepath.Join(home, ".autobox", "configs")
	
	viper.SetDefault("simulation.default_image", "autobox-engine:latest")
	viper.SetDefault("simulation.default_config_path", "/app/configs/simulation.json")
	viper.SetDefault("simulation.default_metrics_path", "/app/configs/metrics.json")
	viper.SetDefault("simulation.default_volumes", []string{
		fmt.Sprintf("%s:/app/configs", defaultConfigsDir),
	})
	viper.SetDefault("simulation.default_environment", map[string]string{})
	viper.SetDefault("simulation.logs_directory", filepath.Join(home, ".autobox", "logs"))
	viper.SetDefault("simulation.configs_directory", defaultConfigsDir)

	viper.SetDefault("output.format", "table")
	viper.SetDefault("output.verbose", false)
	viper.SetDefault("output.color", true)
}

func Get() *Config {
	if cfg == nil {
		if err := Init(); err != nil {
			panic(fmt.Sprintf("failed to initialize config: %v", err))
		}
	}
	return cfg
}

func GetString(key string) string {
	return viper.GetString(key)
}

func GetBool(key string) bool {
	return viper.GetBool(key)
}

func GetInt(key string) int {
	return viper.GetInt(key)
}

func GetStringSlice(key string) []string {
	return viper.GetStringSlice(key)
}

func GetStringMap(key string) map[string]string {
	result := make(map[string]string)
	for k, v := range viper.GetStringMap(key) {
		if str, ok := v.(string); ok {
			result[k] = str
		}
	}
	return result
}