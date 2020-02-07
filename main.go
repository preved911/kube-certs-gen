package main

import (
	// "crypto"
	"fmt"
	"net"
	"os"
	// "github.com/pkg/errors"
	flag "github.com/spf13/pflag"
	"log"
	// "k8s.io/klog"
	certutil "k8s.io/client-go/util/cert"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	// kubeadmcerts "k8s.io/kubernetes/cmd/kubeadm/app/phases/certs"
	// pkiutil "k8s.io/kubernetes/cmd/kubeadm/app/util/pkiutil"
	// "gopkg.in/yaml.v2"
	// "crypto/x509"
	"encoding/json"
	"io/ioutil"
)

func main() {
	var config *string = flag.StringP("config", "c", "", "config file path")

	flag.Parse()

	cfg := getConfig(config)

	if cfg.CertsDir == "" {
		log.Fatalln("[certs] certs store directory should be specified in config file")
	}

	// fmt.Println(cfg)

	fmt.Printf("[certs] Storing certs to: %s\n", cfg.CertsDir)

	// create sa cert and key
	CreateServiceAccountKeyAndPublicKeyFiles(cfg.CertsDir)

	// Create cert list with CA certs
	certList := Certificates{
		&KubeadmCertRootCA,
		&KubeadmCertFrontProxyCA,
		&KubeadmCertEtcdCA,
		// non dinamic certs
		&KubeadmCertKubeletClient,
		&KubeadmCertFrontProxyClient,
		&KubeadmCertEtcdHealthcheck,
		&KubeadmCertEtcdAPIClient,
	}

	(&certList).appendCerts(cfg)

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

func getConfig(configPath *string) CertConfig {
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

	// var cfg []CertConfig
	var cfg CertConfig

	err = json.Unmarshal(config, &cfg)
	if err != nil {
		log.Fatalf("[certs] config unmarshal error: %s\n", err)
	}

	return cfg
}

func (c *Certificates) appendCerts(cfg CertConfig) {
	c.appendEtcdServerCerts(cfg.Etcd.Servers)
	c.appendEtcdPeerCerts(cfg.Etcd.Servers)
	c.appendAPIServerCerts(cfg.APIServer.Servers)
}

func (s *Server) getAltNames() certutil.AltNames {
	var dnsNames []string = []string{"localhost"}
	var ipAddresses []net.IP = []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback}

	for _, value := range s.CertSANs {
		dnsNames = append(dnsNames, value)
	}

	for _, value := range s.CertIPs {
		ip := net.ParseIP(value)
		if ip == nil {
			log.Fatalf("incorrect ip address: %s", value)
		}
		ipAddresses = append(ipAddresses, ip)
	}

	altNames := certutil.AltNames{
		DNSNames: dnsNames,
		IPs:      ipAddresses,
	}

	return altNames
}
