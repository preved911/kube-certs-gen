package main

import (
	"crypto"
	"fmt"
	"github.com/pkg/errors"
	flag "github.com/spf13/pflag"
	// "k8s.io/klog"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	kubeadmcerts "k8s.io/kubernetes/cmd/kubeadm/app/phases/certs"
	pkiutil "k8s.io/kubernetes/cmd/kubeadm/app/util/pkiutil"
)

func main() {
	var certsDir *string = flag.StringP("directory", "d",
		"/etc/kubernetes/pki",
		"certificates will be locate in this directory")
	var etcdHosts *[]string = flag.StringSlice("etcd-host",
		[]string{},
		"list etcd host addresses")

	flag.Parse()

	// klog.V(1).Infof("Storing certs to: %s", *certsDir)
	fmt.Printf("[certs] Storing certs to: %s\n", *certsDir)
	fmt.Println(*etcdHosts)

	// create sa cert and key
	kubeadmcerts.CreateServiceAccountKeyAndPublicKeyFiles(*certsDir)

	// define kubernetes init configuration
	var ic kubeadmapi.InitConfiguration
	ic.ClusterConfiguration.CertificatesDir = *certsDir
	// var etcd kubeadmapi.Etcd
	// etcd.Local = &kubeadmapi.LocalEtcd{
	// 	ServerCertSANs: []string{
	// 		"etcd",
	// 	},
	// }
	// ic.ClusterConfiguration.Etcd = etcd
	// kubeadmcerts.CreatePKIAssets(&ic)
	certTree, err := createPKIAssets(&ic)
	if err != nil {
		fmt.Println(err)
	}

	// for ca, _ := range *certTree {
	// 	fmt.Println(ca)
	// }

	// fmt.Println(ic.ClusterConfiguration.CertificatesDir)
}

// CreatePKIAssets from "k8s.io/kubernetes/cmd/kubeadm/app/phases/certs"
// func createPKIAssets(cfg *kubeadmapi.InitConfiguration) error {
func createPKIAssets(cfg *kubeadmapi.InitConfiguration) (*CertificateTree, error) {
	// fmt.Println(cfg.CertificatesDir)
	// klog.V(1).Infoln("creating PKI assets")
	// klog.V(1).Infoln("Creating CA certs")
	fmt.Println("[certs] Creating CA certs")

	// This structure cannot handle multilevel CA hierarchies.
	// This isn't a problem right now, but may become one in the future.

	// var certList Certificates

	// if cfg.Etcd.Local == nil {
	// 	certList = GetCertsWithoutEtcd()
	// } else {
	// 	certList = GetDefaultCertList()
	// }

	certList := GetDefaultCertList()

	// fmt.Println(certList.AsMap())
	// fmt.Println(certList.AsMap().CertTree())

	certTree, err := certList.AsMap().CertTree()
	if err != nil {
		return nil, err
	}

	// fmt.Println(certTree)
	// for ca, cert := range certTree {
	// for ca, _ := range certTree {
	// 	fmt.Println(ca)
	// 	// fmt.Println(cert)
	// }

	// if err := certTree.CreateTree(cfg); err != nil {
	if err := certTree.createTree(cfg); err != nil {
		return nil, errors.Wrap(err, "error creating PKI assets")
	}

	// fmt.Printf("[certs] Valid certificates and keys now exist in %q\n", cfg.CertificatesDir)

	// // Service accounts are not x509 certs, so handled separately
	// return CreateServiceAccountKeyAndPublicKeyFiles(cfg.CertificatesDir)
	// return nil
	return &certTree, nil
}

func (t CertificateTree) createTree(ic *kubeadmapi.InitConfiguration) error {
	for ca, leaves := range t {
		cfg, err := ca.GetConfig(ic)
		if err != nil {
			return err
		}

		var caKey crypto.Signer

		caCert, err := pkiutil.TryLoadCertFromDisk(ic.CertificatesDir, ca.BaseName)
		if err == nil {
			// Cert exists already, make sure it's valid
			if !caCert.IsCA {
				return errors.Errorf("certificate %q is not a CA", ca.Name)
			}
			// Try and load a CA Key
			caKey, err = pkiutil.TryLoadKeyFromDisk(ic.CertificatesDir, ca.BaseName)
			if err != nil {
				// If there's no CA key, make sure every certificate exists.
				for _, leaf := range leaves {
					cl := certKeyLocation{
						pkiDir:   ic.CertificatesDir,
						baseName: leaf.BaseName,
						uxName:   leaf.Name,
					}
					if err := validateSignedCertWithCA(cl, caCert); err != nil {
						return errors.Wrapf(err, "could not load expected certificate %q or validate the existence of key %q for it", leaf.Name, ca.Name)
					}
				}
				continue
			}
			// CA key exists; just use that to create new certificates.
		} else {
			// CACert doesn't already exist, create a new cert and key.
			caCert, caKey, err = pkiutil.NewCertificateAuthority(cfg)
			if err != nil {
				return err
			}

			err = writeCertificateAuthorityFilesIfNotExist(
				ic.CertificatesDir,
				ca.BaseName,
				caCert,
				caKey,
			)
			if err != nil {
				return err
			}
		}

		// for _, leaf := range leaves {
		// 	if err := leaf.CreateFromCA(ic, caCert, caKey); err != nil {
		// 		return err
		// 	}
		// }
	}
	return nil
}
