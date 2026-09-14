// Package config loads parameters-network service configuration.
package config

import (
	"github.com/jasonmiller-cc/parameters-core/pkg/config"
)

// Config is the top-level configuration for parameters-network.
type Config struct {
	config.BaseConfig `yaml:",inline"`
	Network           NetworkConfig `yaml:"network"`
}

// NetworkConfig holds network-service-specific settings.
type NetworkConfig struct {
	// AllowedInterfaces restricts management to these interface names.
	// An empty list means all interfaces are allowed.
	AllowedInterfaces []string `yaml:"allowed_interfaces"`
	// ManagementInterface is the primary management interface name.
	ManagementInterface string `yaml:"management_interface"`
	// FirewallBackend selects the firewall subsystem: "nftables" or "iptables".
	FirewallBackend string `yaml:"firewall_backend"`
	// NFTablesConfigPath is the path to the nftables configuration file.
	NFTablesConfigPath string `yaml:"nftables_config_path"`
}

const envPrefix = "PARAMS_NETWORK"

// Load reads the YAML config at path and overlays environment variables.
// If path is empty the default location is used.
func Load(path string) (*Config, error) {
	cfg := &Config{}
	if err := config.Load(path, envPrefix, cfg); err != nil {
		return nil, err
	}
	applyDefaults(cfg)
	return cfg, nil
}

func applyDefaults(cfg *Config) {
	if cfg.Network.FirewallBackend == "" {
		cfg.Network.FirewallBackend = "nftables"
	}
	if cfg.Network.NFTablesConfigPath == "" {
		cfg.Network.NFTablesConfigPath = "/etc/nftables.conf"
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8086
	}
}
