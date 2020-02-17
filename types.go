package main

import (
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
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
