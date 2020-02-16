package main

import (
	"fmt"
	"io/ioutil"
	"math"
	"math/big"
	// "os"
	"path/filepath"
	"time"

	"crypto"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"

	"k8s.io/client-go/util/keyutil"
)

func getDataCA() (*x509.Certificate, *rsa.PrivateKey, error) {
	// parse CA key
	caPrivateKeyFile, err := ioutil.ReadFile(filepath.Join(certDir, "ca.key"))
	if err != nil {
		return nil, nil, fmt.Errorf("error read CA private key: %s", err)
	}
	pemBlockKeyCA, _ := pem.Decode(caPrivateKeyFile)
	caKey, err := x509.ParsePKCS1PrivateKey(pemBlockKeyCA.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("error parse CA private key: %s", err)
	}

	// parse CA cert
	caPublicKeyFile, err := ioutil.ReadFile(filepath.Join(certDir, "ca.crt"))
	if err != nil {
		return nil, nil, fmt.Errorf("error read CA public key: %s", err)
	}
	pemBlockCertCA, _ := pem.Decode(caPublicKeyFile)
	caCert, err := x509.ParseCertificate(pemBlockCertCA.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("error parse CA cert: %s", err)
	}

	return caCert, caKey, nil
}

func kubeletCertKeyGen(nodeName string) ([]byte, []byte, error) {
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

	caCert, caKey, err := getDataCA()
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
