package utils

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"time"
)

type CertificateData struct {
	Certificate string
	PrivateKey  string
}

func GenerateCertificate(cert *CertificateData) (tls.Certificate, *CertificateData, error) {
	if cert != nil && cert.Certificate != "" && cert.PrivateKey != "" {
		pair, err := tls.X509KeyPair([]byte(cert.Certificate), []byte(cert.PrivateKey))
		return pair, cert, err
	}
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, nil, err
	}
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Example Co"},
		},
		NotBefore: GetFixedNow().Add(365 * -24 * time.Hour),
		NotAfter:  GetFixedNow().Add(50 * 365 * 24 * time.Hour),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	pub := &priv.PublicKey
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, pub, priv)
	if err != nil {
		return tls.Certificate{}, nil, err
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	privDER, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return tls.Certificate{}, nil, err
	}
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: privDER})
	pair, err := tls.X509KeyPair(certPEM, privPEM)
	return pair, &CertificateData{
		Certificate: string(certPEM),
		PrivateKey:  string(privPEM),
	}, err
}
