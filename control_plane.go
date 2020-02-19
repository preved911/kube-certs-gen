package main

import (
	"crypto"
	"crypto/x509"
	"fmt"
	"github.com/pkg/errors"
	certutil "k8s.io/client-go/util/cert"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	kubeadmcerts "k8s.io/kubernetes/cmd/kubeadm/app/phases/certs"
	pkiutil "k8s.io/kubernetes/cmd/kubeadm/app/util/pkiutil"
	"net"
)

const (
	APIServerDefaultBindAddress = "127.0.0.1"
	APIServerDefaultBindPort    = 6443

	KubeClusterDefaultCIRD = "10.96.0.0/12"
)

func createPKIAssets(cfg *KubeConfig) error {
	cfg.initConfiguration.CertificatesDir = cfg.ClusterConfiguration.CertificatesDir

	fmt.Printf(
		"[kube-certs-gen] Storing certs to: %s\n",
		cfg.initConfiguration.CertificatesDir)

	// define kubernetes init configuration
	cfg.initConfiguration.LocalAPIEndpoint = kubeadmapi.APIEndpoint{
		AdvertiseAddress: APIServerDefaultBindAddress,
		BindPort:         APIServerDefaultBindPort,
	}
	cfg.initConfiguration.Networking = kubeadmapi.Networking{
		ServiceSubnet: KubeClusterDefaultCIRD,
	}

	// create sa cert and key
	err := kubeadmcerts.CreateServiceAccountKeyAndPublicKeyFiles(
		cfg.initConfiguration.CertificatesDir)
	if err != nil {
		return err
	}

	certList := kubeadmcerts.GetDefaultCertList()

	certTree, err := certList.AsMap().CertTree()
	if err != nil {
		return err
	}

	if err := createTree(*cfg, certTree); err != nil {
		return errors.Wrap(err, "error creating PKI assets")
	}

	return nil
}

func createTree(kc KubeConfig, certTree kubeadmcerts.CertificateTree) error {
	ic := kc.initConfiguration

	for ca, leaves := range certTree {
		cfg, err := ca.GetConfig(&ic)
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
				return err
			}
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

		for _, leaf := range leaves {
			if err := createFromCA(leaf, kc, caCert, caKey); err != nil {
				return err
			}
		}
	}

	return nil
}

// CreateFromCA makes and writes a certificate using the given CA cert and key.
func createFromCA(k *kubeadmcerts.KubeadmCert, kc KubeConfig, caCert *x509.Certificate, caKey crypto.Signer) error {
	ic := kc.initConfiguration
	cfg, err := k.GetConfig(&ic)

	if err != nil {
		return errors.Wrapf(err, "couldn't create %q certificate", k.Name)
	}

	switch k.Name {
	case "apiserver":
		err = generateMultipleCertAndKey(k, ic, *cfg, kc.APIServer.Servers, caCert, caKey)
	case "etcd-server":
		err = generateMultipleCertAndKey(k, ic, *cfg, kc.Etcd.Servers, caCert, caKey)
	case "etcd-peer":
		err = generateMultipleCertAndKey(k, ic, *cfg, kc.Etcd.Peers, caCert, caKey)
	default:
		err = generateCertAndKey(k, ic, cfg, caCert, caKey)
	}

	return err
}

func generateMultipleCertAndKey(
	k *kubeadmcerts.KubeadmCert,
	ic kubeadmapi.InitConfiguration,
	cfg certutil.Config,
	servers []Server,
	caCert *x509.Certificate,
	caKey crypto.Signer) error {
	for _, server := range servers {
		altNames := cfg.AltNames
		altNames.DNSNames = append(altNames.DNSNames, server.Certs.SANs...)

		for _, ip := range server.Certs.IPs {
			ip := net.ParseIP(ip)
			if ip == nil {
				return fmt.Errorf("incorrect ip address: %s", ip)
			}

			altNames.IPs = append(altNames.IPs, ip)
		}

		cc := cfg
		cc.AltNames = altNames
		cc.CommonName = server.Name
		cc.AltNames.DNSNames[0] = server.Name
		k.BaseName = fmt.Sprintf("%s-%s", k.BaseName, server.Name)

		if err := generateCertAndKey(k, ic, &cc, caCert, caKey); err != nil {
			return err
		}
	}

	return nil
}

func generateCertAndKey(
	k *kubeadmcerts.KubeadmCert,
	ic kubeadmapi.InitConfiguration,
	cfg *certutil.Config,
	caCert *x509.Certificate,
	caKey crypto.Signer) error {
	cert, key, err := pkiutil.NewCertAndKey(caCert, caKey, cfg)
	if err != nil {
		return err
	}

	err = writeCertificateFilesIfNotExist(
		ic.CertificatesDir,
		k.BaseName,
		caCert,
		cert,
		key,
		cfg,
	)

	if err != nil {
		return errors.Wrapf(err, "failed to write or validate certificate %q", k.Name)
	}

	return nil
}
