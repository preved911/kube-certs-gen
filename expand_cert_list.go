package main

import (
	"fmt"
	kubeadmconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"
)

func (c *Certificates) appendEtcdServerCerts(servers []Server) {
	for _, server := range servers {
		var serverCert *KubeadmCert = &KubeadmCertEtcdServer

		(*serverCert).BaseName = fmt.Sprintf("%s-%s",
			kubeadmconstants.EtcdServerCertAndKeyBaseName,
			server.Name)
		(*serverCert).config.AltNames = server.getAltNames()
		(*serverCert).config.CommonName = server.Name
		(*serverCert).configMutators = []configMutatorsFunc{}

		*c = append(*c, serverCert)
	}
}

func (c *Certificates) appendEtcdPeerCerts(servers []Server) {
	for _, server := range servers {
		var serverCert *KubeadmCert = &KubeadmCertEtcdPeer

		(*serverCert).BaseName = fmt.Sprintf("etcd/peer-%s", server.Name)
		(*serverCert).config.AltNames = server.getAltNames()
		(*serverCert).config.CommonName = server.Name
		(*serverCert).configMutators = []configMutatorsFunc{}

		*c = append(*c, serverCert)
	}
}

func (c *Certificates) appendAPIServerCerts(servers []Server) {
	for _, server := range servers {
		var serverCert *KubeadmCert = &KubeadmCertAPIServer

		(*serverCert).BaseName = fmt.Sprintf("%s-%s",
			kubeadmconstants.APIServerCertAndKeyBaseName,
			server.Name)
		(*serverCert).config.AltNames = server.getAltNames()
		(*serverCert).config.CommonName = server.Name
		(*serverCert).configMutators = []configMutatorsFunc{}

		*c = append(*c, serverCert)
	}
}
