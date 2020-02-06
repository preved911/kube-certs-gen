package main

type certConfig struct {
	Etcd     Etcd
	CertsDir string `json`
}

type Etcd struct {
	Servers []EtcdServer
}

type EtcdServer struct {
	Name     string
	CertSANs []string
	CertIPs  []string
}
