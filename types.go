package main

import (
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
)

// KubeConfig main type for configuration file unmarshaling.
type KubeConfig struct {
	APIServer            APIServer `json:"api" yaml:"api"`
	Etcd                 Etcd
	Nodes                []string             `yaml:"nodes"`
	ClusterConfiguration ClusterConfiguration `yaml:"clusterConfiguration"`
	initConfiguration    kubeadmapi.InitConfiguration
}

// ClusterConfiguration contain cluster configuration fields.
type ClusterConfiguration struct {
	CertificatesDir string `yaml:"certificatesDir"`
}

// APIServer represent control plane apiserver settings.
type APIServer struct {
	Servers []Server
}

// Etcd represent control plane database settings.
type Etcd struct {
	Servers []Server
	Peers   []Server
}

// Server contain any server configuration fields.
type Server struct {
	Name  string
	Certs Cert `yaml:"certs"`
}

// Cert include any server cert valid SANs and IP addresses.
type Cert struct {
	SANs []string `json:"SANs" yaml:"SANs"`
	IPs  []string `json:"IPs" yaml:"IPs"`
}
