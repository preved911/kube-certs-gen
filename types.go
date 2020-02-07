package main

type CertConfig struct {
	CertsDir  string    `json`
	APIServer APIServer `json:"api"`
	Etcd      Etcd
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
