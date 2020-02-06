package main

import (
	// "crypto"
	"fmt"
	"os"
	// "github.com/pkg/errors"
	flag "github.com/spf13/pflag"
	"log"
	// "k8s.io/klog"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	kubeadmcerts "k8s.io/kubernetes/cmd/kubeadm/app/phases/certs"
	// pkiutil "k8s.io/kubernetes/cmd/kubeadm/app/util/pkiutil"
	// "gopkg.in/yaml.v2"
	"encoding/json"
	"io/ioutil"
)

func main() {
	var config *string = flag.StringP("config", "c", "", "config file path")
	// var certsDir *string = flag.StringP("directory", "d",
	// 	"/etc/kubernetes/pki",
	// 	"certificates will be locate in this directory")
	// var etcdHosts *[]string = flag.StringSlice("etcd-host",
	// 	[]string{},
	// 	"list etcd host addresses")

	flag.Parse()

	cfg := getConfig(config)

	if cfg.CertsDir == "" {
		log.Fatalln("[certs] certs store directory should be specified in config file")
	}

	fmt.Println(cfg)

	fmt.Printf("[certs] Storing certs to: %s\n", cfg.CertsDir)

	// create sa cert and key
	kubeadmcerts.CreateServiceAccountKeyAndPublicKeyFiles(cfg.CertsDir)

	// Create cert list with CA certs
	certList := kubeadmcerts.Certificates{
		&kubeadmcerts.KubeadmCertRootCA,
		&kubeadmcerts.KubeadmCertFrontProxyCA,
		&kubeadmcerts.KubeadmCertEtcdCA,
	}

	// Generate yaml file for every host labels
	// and parse it there
	// for _,

	// define kubernetes init configuration
	var ic kubeadmapi.InitConfiguration
	ic.ClusterConfiguration.CertificatesDir = cfg.CertsDir

	certTree, err := certList.AsMap().CertTree()
	if err != nil {
		log.Fatalf("unexpected error occured: %s", err)
	}

	if err := certTree.CreateTree(&ic); err != nil {
		log.Fatalf("error creating PKI assets: %s", err)
	}
}

func getConfig(configPath *string) certConfig {
	if *configPath == "" {
		log.Fatalln("config path cannot be empty")
	}

	configFile, err := os.Open(*configPath)
	if err != nil {
		log.Fatalf("[certs] config read error: %s\n", err)
	}
	defer configFile.Close()
	config, err := ioutil.ReadAll(configFile)
	if err != nil {
		log.Fatalf("[certs] config read error: %s\n", err)
	}

	// var cfg []certConfig
	var cfg certConfig

	err = json.Unmarshal(config, &cfg)
	if err != nil {
		log.Fatalf("[certs] config unmarshal error: %s\n", err)
	}

	return cfg
}
