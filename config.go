package main

import (
	"fmt"
	"net"
	"os"

	yaml "gopkg.in/yaml.v3"
)

func LoadConfig(path string) (*Config, error) {

	file, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// initialize config
	// &, yes, it is pointer
	// Right now, in our business, we don't have to modify our config file after we load it.
	// So why we need to pointer here?
	// 1. yaml.Unmarshal required pointer to write data into it
	// 2. LoadConfig returns *Config, so declaring as pointer from the start is convenient
	// We can init like this, but current way is more like Go idiom xDD
	/*
		var config Config
		err = yaml.Unmarshal(file, &config)
		// ...
		return &config, nil
	*/
	config := &Config{}

	// unmarshal yaml
	err = yaml.Unmarshal(file, config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal yaml: %w", err)
	}

	// Set default ssh port if not define
	for i := range config.Hosts {
		if config.Hosts[i].SSH_Port == 0 {
			config.Hosts[i].SSH_Port = 22
		}
	}

	// Validate IP
	for _, host := range config.Hosts {
		if net.ParseIP(host.IP) == nil {
			return nil, fmt.Errorf("invalid IP address: %s", host.IP)
		}
	}

	return config, nil
}
