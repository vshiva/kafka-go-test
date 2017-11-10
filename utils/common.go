package utils

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
)

//NewTLSConfig - New TLS Config that can help setting up HTTPS Server / Client
func NewTLSConfig(certFile, keyFile, caFile string, insecureSkipVerify bool) (t *tls.Config, err error) {

	if certFile != "" && keyFile != "" && caFile != "" {
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			log.Fatal(err)
		}

		caCert, err := ioutil.ReadFile(caFile)
		if err != nil {
			log.Fatal(err)
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		t = &tls.Config{
			Certificates:       []tls.Certificate{cert},
			RootCAs:            caCertPool,
			InsecureSkipVerify: insecureSkipVerify,
		}
	}
	// will be nil by default if nothing is provided
	return t, err
}
