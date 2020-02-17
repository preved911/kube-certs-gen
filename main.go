package main

import (
	flag "github.com/spf13/pflag"
	"log"
)

func main() {
	var configPath *string = flag.StringP("config", "c", "", "config file path")

	flag.Parse()

	cfg, err := parseConfig(configPath)
	if err != nil {
		log.Fatalln(err)
	}

	if err := createPKIAssets(cfg); err != nil {
		log.Fatalln(err)
	}

	if err := kubeletConfigCreate("fr-node-detector-0", cfg.initConfiguration.CertificatesDir); err != nil {
		log.Fatalln(err)
	}
}
