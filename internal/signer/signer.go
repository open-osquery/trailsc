// Package signer package contains utilities for signing a given blob and return
// the certificate file and related intermediate certificates to be bundled with
// the package it's signing to be able to verify at the remote end.
package signer

import (
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"io"
	"io/ioutil"
	"strings"

	"github.com/pkg/errors"
)

// ParseCertificates takes a reader, usually a file with stacked PEM blocks in
// PKCS1 format that contains the private key and certificates. The certificates
// must be in the following order, leaf, intermediate..., [root]. The root CA is
// not required here.
// Note: The private key in the certificate file must be decrypted.
func ParseCertificates(certFile io.Reader) (
	signer crypto.Signer, leaf []byte, intermediate [][]byte, err error,
) {
	if certFile == nil {
		return nil, nil, nil, errors.New("Failed to read certs")
	}

	var certs []*pem.Block
	rawCert, _ := ioutil.ReadAll(certFile)
	for {
		var p *pem.Block
		p, rawCert = pem.Decode(rawCert)
		if p == nil {
			break
		}

		if p.Type == "CERTIFICATE" {
			certs = append(certs, p)
		} else if strings.HasSuffix(p.Type, "PRIVATE KEY") {
			signer, err = x509.ParsePKCS1PrivateKey(p.Bytes)
			if err != nil {
				return nil, nil, nil, errors.Wrap(
					err, "Failed to parse (encrypted?) private key)")
			}
		}
	}

	if len(certs) > 1 {
		leaf = pem.EncodeToMemory(certs[0])
	}

	for i := 1; i < len(certs); i++ {
		c, err := x509.ParseCertificate(certs[i].Bytes)
		if err != nil {
			continue
		}

		if c.Issuer.String() == c.Subject.String() {
			// Self signed CA certificate, omit from the list
			continue
		}

		intermediate = append(intermediate, pem.EncodeToMemory(certs[i]))
	}

	return signer, leaf, intermediate, nil
}
