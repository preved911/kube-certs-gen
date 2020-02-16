package main

import (
	"encoding/json"
	"errors"
	"fmt"
	// flag "github.com/spf13/pflag"
	"io/ioutil"
	certutil "k8s.io/client-go/util/cert"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	"log"
	"net"
	"os"
	// "encoding/pem"
)

type KubeConfig struct {
	APIServer            APIServer `json:"api"`
	Etcd                 Etcd
	ClusterConfiguration kubeadmapi.ClusterConfiguration
	// LocalAPIEndpoint     kubeadmapi.APIEndpoint `json:",omitempty"`
	initConfiguration kubeadmapi.InitConfiguration `json:"-"`
}

type APIServer struct {
	Servers []Server
}

type Etcd struct {
	Servers []Server
	Peers   []Server
}

type Server struct {
	Name     string
	CertSANs []string
	CertIPs  []string
}

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
