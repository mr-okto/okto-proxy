package certificates

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"io/ioutil"
	"log"
	"math/big"
	"net"
	"time"
)

const RootCert = "certs/ca.crt"
const RootKey  = "certs/ca.key"

var rootCert tls.Certificate

func LoadRootCert() error {
	certFd, err := ioutil.ReadFile(RootCert)
	if err != nil {
		return err
	}
	keyFd, err := ioutil.ReadFile(RootKey)
	if err != nil {
		return err
	}
	rootCert, err = tls.X509KeyPair(certFd, keyFd)
	if err != nil {
		return err
	}
	rootCert.Leaf, err = x509.ParseCertificate(rootCert.Certificate[0])
	return err
}

func GetCert(hosts ...string) (*tls.Certificate, error) {
	key, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	certSerial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	certConfig := &x509.Certificate{
		SerialNumber: certSerial,
		Subject:      pkix.Name{CommonName: hosts[0]},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Hour * 24),
		KeyUsage: x509.KeyUsageDigitalSignature |
			x509.KeyUsageContentCommitment |
			x509.KeyUsageKeyEncipherment |
			x509.KeyUsageDataEncipherment |
			x509.KeyUsageKeyAgreement |
			x509.KeyUsageCertSign |
			x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		SignatureAlgorithm:    x509.SHA256WithRSA,
	}
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			certConfig.IPAddresses = append(certConfig.IPAddresses, ip)
		} else {
			certConfig.DNSNames = append(certConfig.DNSNames, h)
		}
	}
	leaf, err := x509.CreateCertificate(rand.Reader, certConfig,
		rootCert.Leaf, key.Public(), rootCert.PrivateKey)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	res, err := x509.ParseCertificate(leaf)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	cert := &tls.Certificate{
		Certificate: [][]byte{leaf},
		PrivateKey:  key,
		Leaf:        res,
	}
	return cert, nil

}
