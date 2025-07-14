package utils

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"log"
	"os"
)

func decodeAndWriteToFile(envVar, destPath string) error {
	b64 := os.Getenv(envVar)
	if b64 == "" {
		return fmt.Errorf("missing env var: %s", envVar)
	}
	data, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return fmt.Errorf("failed to decode %s: %w", envVar, err)
	}
	return os.WriteFile(destPath, data, 0600)
}


func Decode()(tls.Certificate,*x509.CertPool){
	err := decodeAndWriteToFile("SERVICE_CERT_BASE64", "/tmp/service.cert")
	if err != nil {
		log.Fatalf("cert write error: %v", err)
	}
	err = decodeAndWriteToFile("SERVICE_KEY_BASE64", "/tmp/service.key")
	if err != nil {
		log.Fatalf("key write error: %v", err)
	}
	err = decodeAndWriteToFile("CA_PEM_BASE64", "/tmp/ca.pem")
	if err != nil {
		log.Fatalf("ca write error: %v", err)
	}

	keypair, err := tls.LoadX509KeyPair("/tmp/service.cert", "/tmp/service.key")
	if err != nil {
		log.Fatalf("Failed to load TLS keypair: %v", err)
	}

	caCert, err := os.ReadFile("/tmp/ca.pem")
	if err != nil {
		log.Fatalf("Failed to read CA cert: %v", err)
	}
	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		log.Fatalf("Failed to parse CA PEM")
	}
	return keypair,caCertPool
} 