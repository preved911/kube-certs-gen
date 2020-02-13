package main

import (
	"fmt"
	"os"
	"time"
	// "errors"

	// "k8s.io/client-go/util/keyutil"

	"crypto"

	// "crypto/ecdsa"
	// "crypto/elliptic"

	cryptorand "crypto/rand"

	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"

	// certutil "k8s.io/client-go/util/cert"

	"k8s.io/client-go/util/keyutil"
	// "k8s.io/apimachinery/pkg/types"
	// bootstrap "k8s.io/kubernetes/pkg/kubelet/certificate/bootstrap"
	"io/ioutil"

	"math"
	"math/big"
)

func generateKubeletConfig(nodeName string) error {
	// privKeyPath =
	// keyData, _, err = keyutil.LoadOrGenerateKeyFile(privKeyPath)
	// privateKey, err := ecdsa.GenerateKey(elliptic.P256(), cryptorand.Reader)
	// fmt.Println(privateKey)

	// keyData, _ := keyutil.MakeEllipticPrivateKeyPEM()
	// block, _ := pem.Decode([]byte(string(keyData)))
	// x509Encoded := block.Bytes
	// privateKey, _ := x509.ParseECPrivateKey(x509Encoded)
	// fmt.Println(privateKey)

	// p, _, _ := keyutil.LoadOrGenerateKeyFile("./test.key")
	// p1, _ := keyutil.ParsePrivateKeyPEM(p)
	// fmt.Println(p1)

	// var nodeName types.NodeName = "node-1"

	// err := bootstrap.LoadClientCert(

	// err := LoadClientCert(
	// 	"/home/q/Downloads/exercism/go/kube-certs-gen/kubernetes/kubelet.conf",
	// 	"/home/q/Downloads/exercism/go/kube-certs-gen/kubernetes/bootstrap-kubelet.conf",
	// 	"/home/q/Downloads/exercism/go/kube-certs-gen/kubernetes/certs",
	// 	nodeName,
	// )
	// fmt.Println(err)

	// create private key
	privateKeyData, err := keyutil.MakeEllipticPrivateKeyPEM()
	if err != nil {
		return fmt.Errorf("error generating key: %v", err)
	}
	key, err := keyutil.ParsePrivateKeyPEM(privateKeyData)
	if err != nil {
		return fmt.Errorf("invalid private key for certificate request: %v", err)
	}
	fmt.Println(string(privateKeyData))

	subject := &pkix.Name{
		Organization: []string{"system:nodes"},
		CommonName:   "system:node:" + string(nodeName),
	}

	// var cfg certutil.Config = &certutil.Config{
	// 	CommonName:   "system:node:" + string(nodeName),
	// 	Organization: []string{"system:nodes"},

	// crease CSR
	// csrData, err := certutil.MakeCSR(privateKey, subject, nil, nil)
	// if err != nil {
	// 	return fmt.Errorf("unable to generate certificate request: %v", err)
	// }
	// pemBlockCSR, err := pem.Decode(csrData)
	// if err != nil {
	// 	return fmt.Errorf("error decode certificate request: %v", err)
	// }
	// clientCSR, err := x509.ParseCertificateRequest(pemBlockCSR)
	// if err != nil {
	// 	return fmt.Errorf("error parse certificate request: %v", err)
	// }

	// parse CA cert
	caPrivateKeyFile, err := ioutil.ReadFile("/home/q/Downloads/exercism/go/kube-certs-gen/kubernetes/pki/ca.key")
	if err != nil {
		return fmt.Errorf("error read CA private key: %s", err)
	}
	// pemBlockKeyCA, err := pem.Decode(caPrivateKeyFile)
	pemBlockKeyCA, _ := pem.Decode(caPrivateKeyFile)
	// if err != nil {
	// 	return fmt.Errorf("error decode CA private key: %s", err)
	// }

	// der, err := x509.DecryptPEMBlock(pemBlock, []byte("ca private key password"))
	// if err != nil {
	//     panic(err)
	// }
	// caPrivateKey, err := x509.ParsePKCS1PrivateKey(der)
	caKey, err := x509.ParsePKCS1PrivateKey(pemBlockKeyCA.Bytes)
	if err != nil {
		return fmt.Errorf("error parse CA private key: %s", err)
	}

	caPublicKeyFile, err := ioutil.ReadFile("/home/q/Downloads/exercism/go/kube-certs-gen/kubernetes/pki/ca.crt")
	if err != nil {
		return fmt.Errorf("error read CA public key: %s", err)
	}
	pemBlockCertCA, _ := pem.Decode(caPublicKeyFile)
	// if err != nil {
	// 	return fmt.Errorf("error decode CA cert: %s", err)
	// }
	caCert, err := x509.ParseCertificate(pemBlockCertCA.Bytes)
	if err != nil {
		return fmt.Errorf("error parse CA cert: %s", err)
	}

	serial, err := cryptorand.Int(cryptorand.Reader, new(big.Int).SetInt64(math.MaxInt64))
	if err != nil {
		return fmt.Errorf("error generate serial: %s", err)
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
	// keySigner, ok := privateKeyData.(crypto.Signer)
	if !ok {
		return fmt.Errorf("private key does not implement crypto.Signer")
	}

	certDERBytes, err := x509.CreateCertificate(cryptorand.Reader, &certTemplate, caCert, keySigner.Public(), caKey)
	if err != nil {
		return fmt.Errorf("error generate certificate: %s", err)
	}

	cert, err := x509.ParseCertificate(certDERBytes)

	block := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDERBytes,
	}

	fmt.Println(pem.Decode(block.Bytes))
	fmt.Println(cert)

	clientCRTFile, err := os.Create("bob.crt")
	if err != nil {
		panic(err)
	}
	pem.Encode(clientCRTFile, &pem.Block{Type: "CERTIFICATE", Bytes: certDERBytes})
	clientCRTFile.Close()

	return nil
}
