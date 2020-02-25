package main

import (
	"fmt"
	"io/ioutil"
	"math"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"crypto"
	"crypto/ecdsa"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"

	"k8s.io/client-go/util/keyutil"

	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	kubeletphase "k8s.io/kubernetes/cmd/kubeadm/app/phases/kubelet"
	// kubeletconfig "k8s.io/kubernetes/pkg/kubelet/apis/config"
	kubeletconfig "k8s.io/kubelet/config/v1beta1"
)

// func parseCertCA(certificatesDir string) ([]byte, error) {
func parseCertOrKeyCA(certificatesDir, fileName string) (*pem.Block, error) {
	ca, err := ioutil.ReadFile(filepath.Join(certificatesDir, fileName))
	if err != nil {
		return nil, fmt.Errorf("error read %s file: %s", fileName, err)
	}

	pemBlock, _ := pem.Decode(ca)

	return pemBlock, nil
}

func getDataCA(certificatesDir string) (*x509.Certificate, *rsa.PrivateKey, error) {
	// parse CA key
	pemBlockKeyCA, err := parseCertOrKeyCA(certificatesDir, "ca.key")
	if err != nil {
		return nil, nil, err
	}

	caKey, err := x509.ParsePKCS1PrivateKey(pemBlockKeyCA.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("error parse CA private key: %s", err)
	}

	pemBlockCertCA, err := parseCertOrKeyCA(certificatesDir, "ca.crt")
	if err != nil {
		return nil, nil, err
	}

	caCert, err := x509.ParseCertificate(pemBlockCertCA.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("error parse CA cert: %s", err)
	}

	return caCert, caKey, nil
}

// func kubeletCertKeyGen(nodeName, certificatesDir string) ([]byte, []byte, error) {
func kubeletCertKeyGen(nodeName, certificatesDir string) (*pem.Block, *pem.Block, error) {
	// create private key
	privateKeyData, err := keyutil.MakeEllipticPrivateKeyPEM()
	if err != nil {
		return nil, nil, fmt.Errorf("error generating key: %v", err)
	}

	privateKey, err := keyutil.ParsePrivateKeyPEM(privateKeyData)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid private key for certificate request: %v", err)
	}

	subject := &pkix.Name{
		Organization: []string{"system:nodes"},
		CommonName:   "system:node:" + nodeName,
	}

	caCert, caKey, err := getDataCA(certificatesDir)
	if err != nil {
		return nil, nil, err
	}

	serial, err := cryptorand.Int(cryptorand.Reader, new(big.Int).SetInt64(math.MaxInt64))
	if err != nil {
		return nil, nil, fmt.Errorf("error generate serial: %s", err)
	}

	// Generate signed cert
	certTemplate := x509.Certificate{
		Subject:      *subject,
		SerialNumber: serial,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageClientAuth,
		},
	}

	keySigner, ok := privateKey.(crypto.Signer)
	if !ok {
		return nil, nil, fmt.Errorf("private key does not implement crypto.Signer")
	}

	publicKey, err := x509.CreateCertificate(cryptorand.Reader, &certTemplate, caCert, keySigner.Public(), caKey)
	if err != nil {
		return nil, nil, fmt.Errorf("error generate certificate: %s", err)
	}

	publicKeyBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: publicKey,
	}

	key, ok := privateKey.(*ecdsa.PrivateKey)
	if !ok {
		return nil, nil, fmt.Errorf("error transform private key interface{} to *ecdsa.PrivateKey")
	}

	keyBytes, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil, nil, fmt.Errorf("error marshal private key: %s", err)
	}

	privateKeyBlock := &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: keyBytes,
	}

	return publicKeyBlock, privateKeyBlock, nil
}

func writeKubeletClientPem(nodeName, certificatesDir string) error {
	var kubeletClientPemFile string = fmt.Sprintf("kubelet-client-%s.pem", nodeName)

	// check kubelet.conf existence
	kubeletInfo, err := os.Stat(filepath.Join(certificatesDir, kubeletClientPemFile))
	if err == nil && kubeletInfo.Size() > 0 && !kubeletInfo.IsDir() {
		fmt.Printf("[kube-certs-gen] Using the existing \"kubelet-client-%s.pem\" from disk\n", nodeName)
		return nil
	}

	// generate kubelet client cert and key
	cert, key, err := kubeletCertKeyGen(nodeName, certificatesDir)
	if err != nil {
		return err
	}

	pemOutFile, err := os.Create(
		filepath.Join(
			certificatesDir,
			fmt.Sprintf("kubelet-client-%s.pem", nodeName),
		),
	)
	if err != nil {
		return fmt.Errorf("Failed to open \"kubelet-client-%s.pem\" for writing: %s", nodeName, err)
	}

	defer pemOutFile.Close()

	if err := pem.Encode(pemOutFile, cert); err != nil {
		return fmt.Errorf("Failed to write public key data to \"kubelet-client-%s.pem\": %s", nodeName, err)
	}

	// fmt.Printf("[kube-cert-gen] Writing public key data to kubelet-client-%s.pem\n", nodeName)

	if err := pem.Encode(pemOutFile, key); err != nil {
		return fmt.Errorf("Failed to write private key data to \"kubelet-client-%s.pem\": %s", nodeName, err)
	}

	// fmt.Printf("[kube-cert-gen] Writing private key data to kubelet-client-%s.pem\n", nodeName)
	fmt.Printf("[kube-cert-gen] Writing kubelet client pem data to \"kubelet-client-%s.pem\"\n", nodeName)

	return nil
}

func kubeletKubeConfigCreate(certificatesDir string) error {
	caCertData, err := parseCertOrKeyCA(certificatesDir, "ca.crt")
	if err != nil {
		return err
	}

	clientConfig := &restclient.Config{
		Host: "https://127.0.0.1:6443",
	}

	// Build resulting kubeconfig.
	kubeconfigData := clientcmdapi.Config{
		// Define a cluster stanza based on the bootstrap kubeconfig.
		Clusters: map[string]*clientcmdapi.Cluster{"default-cluster": {
			Server:                   clientConfig.Host,
			CertificateAuthorityData: caCertData.Bytes,
		}},
		// Define auth based on the obtained client cert.
		AuthInfos: map[string]*clientcmdapi.AuthInfo{"default-auth": {
			ClientCertificate: "/var/lib/kubelet/pki/kubelet-client-current.pem",
			ClientKey:         "/var/lib/kubelet/pki/kubelet-client-current.pem",
		}},
		// Define a context that connects the auth info and cluster, and set it as the default
		Contexts: map[string]*clientcmdapi.Context{"default-context": {
			Cluster:   "default-cluster",
			AuthInfo:  "default-auth",
			Namespace: "default",
		}},
		CurrentContext: "default-context",
	}

	fmt.Println("[kube-certs-gen] Write \"kubelet.conf\" to disk")

	// Marshal to disk
	return clientcmd.WriteToFile(
		kubeconfigData,
		filepath.Join(certificatesDir, "kubelet.conf"),
	)
}

func kubeletConfigCreate(certificatesDir string) error {
	var (
		healthzPort                                                                int32 = 10248
		kubeletAnonymousAuthenticationEnabled, kubeletWebhookAuthenticationEnabled bool
	)

	kubeletWebhookAuthenticationEnabled = true

	kubeletConfig := &kubeletconfig.KubeletConfiguration{
		Authentication: kubeletconfig.KubeletAuthentication{
			X509: kubeletconfig.KubeletX509Authentication{
				ClientCAFile: "/etc/kubernetes/pki/ca.crt",
			},
			Webhook: kubeletconfig.KubeletWebhookAuthentication{
				Enabled: &kubeletWebhookAuthenticationEnabled,
			},
			Anonymous: kubeletconfig.KubeletAnonymousAuthentication{
				Enabled: &kubeletAnonymousAuthenticationEnabled,
			},
		},
		Authorization: kubeletconfig.KubeletAuthorization{
			Mode: "Webhook",
		},
		CgroupDriver:       "systemd",
		ClusterDNS:         []string{"10.96.0.10"},
		ClusterDomain:      "cluster.local",
		HealthzBindAddress: "127.0.0.1",
		HealthzPort:        &healthzPort,
		RotateCertificates: true,
		StaticPodPath:      "/etc/kubernetes/manifests",
	}

	return kubeletphase.WriteConfigToDisk(
		kubeletConfig,
		certificatesDir)
}
