package darp

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"time"
)

type CACerts struct {
	CA           *x509.Certificate
	CAPrivateKey *rsa.PrivateKey
	CAPem        []byte
	CAPrivPem    []byte
}

func (ca *CACerts) generateRootCerts() (err error) {
	// Create root certificate template
	ca.CA = &x509.Certificate{
		SerialNumber:          big.NewInt(2019),
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(100, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
	// Generate root private key
	ca.CAPrivateKey, err = rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		log.Error(err, "failed to generate CA PRIVATE KEY")
		return err
	}
	// Create root certificate
	caBytes, err := x509.CreateCertificate(rand.Reader, ca.CA, ca.CA, &ca.CAPrivateKey.PublicKey, ca.CAPrivateKey)
	if err != nil {
		log.Error(err, "failed to generate CA crt ")
		return err
	}
	// Encode ca public key into base64 byte array
	caPEM := new(bytes.Buffer)
	err = pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})

	if err != nil {
		log.Error(err, "failed to encode CA PEM to base64")
		return err
	}
	// Encode ca private key into base64 byte array
	certPrivKeyPEM := new(bytes.Buffer)
	err = pem.Encode(certPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(ca.CAPrivateKey),})
	if err != nil {
		log.Error(err, "failed to encode key")
		return err
	}

	// Read from buffers
	ca.CAPem = caPEM.Bytes()
	ca.CAPrivPem = certPrivKeyPEM.Bytes()
	return nil
}

func (ca *CACerts) generateCertificates(serviceName string) (crt []byte, key []byte, err error) {
	log.Info("Certificate OU", "ou", serviceName)
	// Create certificate
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(1658),
		Subject:      pkix.Name{CommonName: serviceName,},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
	// Generate private key
	certPrivKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		log.Error(err, "failed to generate certPrivKey for client certificate")
		return nil, nil, err
	}
	// Generate certificate
	certBytes, err := x509.CreateCertificate(rand.Reader, cert, ca.CA, &certPrivKey.PublicKey, ca.CAPrivateKey)
	if err != nil {
		log.Error(err, "failed to create certificate")
		return nil, nil, err
	}
	// Encode certificate to base64 byte array
	certPEM := new(bytes.Buffer)
	err = pem.Encode(certPEM, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes,})
	if err != nil {
		log.Error(err, "failed to encode certificate")
		return nil, nil, err
	}
	// Encode key to base64 byte array
	certPrivKeyPEM := new(bytes.Buffer)
	err = pem.Encode(certPrivKeyPEM, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(certPrivKey),})
	if err != nil {
		log.Error(err, "failed to encode key")
		return nil, nil, err
	}
	// Read from buffers and return results
	crt = certPEM.Bytes()
	key = certPrivKeyPEM.Bytes()
	return crt, key, nil
}

func (ca *CACerts) loadRootCertificates() (err error) {

	caBlock, _ := pem.Decode(ca.CAPem)
	keyBlock, _ := pem.Decode(ca.CAPrivPem)
	ca.CA, err = x509.ParseCertificate(caBlock.Bytes)
	if err != nil {
		log.Error(err, "failed to parse root crt")
		return err
	}
	ca.CAPrivateKey, err = x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil {
		log.Error(err, "failed to parse root key")
		return err
	}
	return nil
}

func (ca *CACerts) validateCertificates() {
	// First, create the set of root certificates. For this example we only
	// have one. It's also possible to omit this in order to use the
	// default root set of the current operating system.
	//roots := x509.NewCertPool()
	//ok := roots.AppendCertsFromPEM([]byte(rootPEM))
	//if !ok {
	//	panic("failed to parse root certificate")
	//}
	//
	//block, _ := pem.Decode([]byte(certPEM))
	//if block == nil {
	//	panic("failed to parse certificate PEM")
	//}
	//cert, err := x509.ParseCertificate(block.Bytes)
	//if err != nil {
	//	panic("failed to parse certificate: " + err.Error())
	//}
	//
	//opts := x509.VerifyOptions{
	//	DNSName: "mail.google.com",
	//	Roots:   roots,
	//}
	//
	//if _, err := cert.Verify(opts); err != nil {
	//	panic("failed to verify certificate: " + err.Error())
	//}
}
