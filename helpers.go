package main

import (
	"encoding/json"
	"errors"
	"fmt"
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
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

	switch filepath.Ext(*configPath) {
	case ".json":
		if err = json.Unmarshal(config, &cfg); err != nil {
			return nil, fmt.Errorf("config unmarshal error: %s", err)
		}
	case ".yaml":
		if err = yaml.Unmarshal([]byte(config), &cfg); err != nil {
			return nil, fmt.Errorf("config unmarshal error: %s", err)
		}
	default:
		return nil, errors.New("incorrect file extension")
	}

	return &cfg, nil
}
