package main

import (
	// "encoding/json"
	"fmt"
	flag "github.com/spf13/pflag"
	// "io/ioutil"
	// certutil "k8s.io/client-go/util/cert"
	// kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	// kubeadmcerts "k8s.io/kubernetes/cmd/kubeadm/app/phases/certs"
	// "log"
	// "net"
	// "encoding/pem"
	// "os"
)

var certDir string

func main() {
	var config *string = flag.StringP("config", "c", "", "config file path")

	flag.Parse()

	err := createPKIAssets(config)
	if err != nil {
		fmt.Println(err)
	}

	// cfg, err := getConfig(config)
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// fmt.Printf("[kube-certs-gen] Storing certs to: %s\n", cfg.CertsDir)

	// // create sa cert and key
	// // CreateServiceAccountKeyAndPublicKeyFiles(cfg.CertsDir)

	// // define kubernetes init configuration
	// var ic kubeadmapi.InitConfiguration
	// ic.ClusterConfiguration.CertificatesDir = cfg.CertsDir
	// ic.LocalAPIEndpoint = kubeadmapi.APIEndpoint{
	// 	AdvertiseAddress: "127.0.0.1",
	// 	BindPort:         6443,
	// }
	// ic.Networking = kubeadmapi.Networking{
	// 	ServiceSubnet: "10.96.0.0/12",
	// }

	// ca := &kubeadmcerts.KubeadmCertRootCA
	// ca := &kubeadmcerts.KubeadmCertEtcdServer
	// ca := &kubeadmcerts.KubeadmCertAPIServer
	// c, err := ca.GetConfig(&ic)
	// if err != nil {
	// 	log.Fatalln(err)
	// }
	// fmt.Println(*c)

	// certTree, err := certList.AsMap().CertTree()
	// if err != nil {
	// 	log.Fatalf("[kube-certs-gen] unexpected error occured: %s", err)
	// }

	// if err := certTree.CreateTree(&ic); err != nil {
	// 	log.Fatalf("[kube-certs-gen] error creating PKI assets: %s", err)
	// }

	// writePrivKey()
	// caCert, caKey, err := KubeletConfigCreate("node-1")
	// caCert, caKey, err := kubeletCertKeyGen("node-1")
	// fmt.Println(err)
	// kubeletPem, err := os.Create("certs/kubelet.pem")
	// fmt.Println(err)
	// defer kubeletPem.Close()
	// pem.Encode(kubeletPem, &pem.Block{Type: "CERTIFICATE", Bytes: caCert})
	// pem.Encode(kubeletPem, &pem.Block{Type: "EC PRIVATE KEY", Bytes: caKey})
}
