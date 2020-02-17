package main

import (
	"fmt"
	flag "github.com/spf13/pflag"
	"log"
	"os"
)

const AppVersion = "0.0.11"

func main() {
	var configPath *string = flag.StringP("config", "c", "", "specify config file path")
	var version *bool = flag.BoolP("version", "v", false, "return tool version and exit")

	flag.Parse()

	if *version {
		fmt.Printf("kube-certs-gen, version %s\n", AppVersion)
		os.Exit(0)
	}

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
