package main

import (
	"fmt"
	"io/ioutil"
	"math"
	"math/big"
	"path/filepath"
	"time"

	"crypto"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"

	"k8s.io/client-go/util/keyutil"

	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
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

func kubeletCertKeyGen(nodeName, certificatesDir string) ([]byte, []byte, error) {
	// create private key
	privateKeyData, err := keyutil.MakeEllipticPrivateKeyPEM()
	if err != nil {
		return nil, nil, fmt.Errorf("error generating key: %v", err)
	}
	key, err := keyutil.ParsePrivateKeyPEM(privateKeyData)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid private key for certificate request: %v", err)
	}

	subject := &pkix.Name{
		Organization: []string{"system:nodes"},
		CommonName:   "system:node:" + string(nodeName),
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

	keySigner, ok := key.(crypto.Signer)
	if !ok {
		return nil, nil, fmt.Errorf("private key does not implement crypto.Signer")
	}

	publicKey, err := x509.CreateCertificate(cryptorand.Reader, &certTemplate, caCert, keySigner.Public(), caKey)
	if err != nil {
		return nil, nil, fmt.Errorf("error generate certificate: %s", err)
	}

	return publicKey, privateKeyData, nil
}

func kubeletConfigCreate(nodeName, certificatesDir string) error {
	cert, key, err := kubeletCertKeyGen(nodeName, certificatesDir)
	if err != nil {
		return err
	}

	caCertData, err := parseCertOrKeyCA(certificatesDir, "ca.crt")
	if err != nil {
		return err
	}

	clientConfig := &restclient.Config{
		Host: "https://127.0.0.1:6443",
	}

	certData := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert})
	keyData := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: key})

	// Build resulting kubeconfig.
	kubeconfigData := clientcmdapi.Config{
		// Define a cluster stanza based on the bootstrap kubeconfig.
		Clusters: map[string]*clientcmdapi.Cluster{"default-cluster": {
			Server:                   clientConfig.Host,
			CertificateAuthorityData: caCertData.Bytes,
		}},
		// Define auth based on the obtained client cert.
		AuthInfos: map[string]*clientcmdapi.AuthInfo{"default-auth": {
			ClientCertificateData: certData,
			ClientKeyData:         keyData,
		}},
		// Define a context that connects the auth info and cluster, and set it as the default
		Contexts: map[string]*clientcmdapi.Context{"default-context": {
			Cluster:   "default-cluster",
			AuthInfo:  "default-auth",
			Namespace: "default",
		}},
		CurrentContext: "default-context",
	}

	fmt.Printf("[kube-certs-gen] Write \"kubelet-%s.conf\" to disk\n", nodeName)

	// Marshal to disk
	return clientcmd.WriteToFile(
		kubeconfigData,
		filepath.Join(certificatesDir, fmt.Sprintf("kubelet-%s.conf", nodeName)),
	)
}
