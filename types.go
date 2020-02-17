package main

import (
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
)

type KubeConfig struct {
	APIServer            APIServer `json:"api" yaml:"api"`
	Etcd                 Etcd
	ClusterConfiguration ClusterConfiguration         `yaml:"clusterConfiguration"`
	initConfiguration    kubeadmapi.InitConfiguration `json:"-" yaml:"-"`
}

type ClusterConfiguration struct {
	CertificatesDir string `yaml:"certificatesDir"`
}

type APIServer struct {
	Servers []Server
}

type Etcd struct {
	Servers []Server
	Peers   []Server
}

type Server struct {
	Name  string
	Certs Cert `yaml:"certs"`
}

type Cert struct {
	SANs []string `json:"SANs" yaml:"SANs"`
	IPs  []string `json:"IPs" yaml:"IPs"`
}
