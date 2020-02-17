package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
)

func parseConfig(configPath *string) (*KubeConfig, error) {
	if *configPath == "" {
		return nil, errors.New("config path cannot be empty")
	}

	configFile, err := os.Open(*configPath)
	if err != nil {
		return nil, fmt.Errorf("config read error: %s", err)
	}
	defer configFile.Close()
	config, err := ioutil.ReadAll(configFile)
	if err != nil {
		return nil, fmt.Errorf("config read error: %s", err)
	}

	var cfg KubeConfig

	err = json.Unmarshal(config, &cfg)
	if err != nil {
		return nil, fmt.Errorf("config unmarshal error: %s", err)
	}

	return &cfg, nil
}
