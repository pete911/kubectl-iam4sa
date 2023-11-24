package aws

import (
	"crypto/sha1"
	"crypto/tls"
	"fmt"
	"net"
	"net/url"
	"time"
)

// FingerprintSHA1 returns certificate sha1 fingerprint e.g. oidc.eks.eu-west-2.amazonaws.com
func FingerprintSHA1(addr string, tlsSkipVerify bool) (string, error) {
	hostAndPort, err := getHostAndPort(addr)
	if err != nil {
		return "", err
	}

	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: 5 * time.Second}, "tcp", hostAndPort, &tls.Config{InsecureSkipVerify: tlsSkipVerify})
	if err != nil {
		return "", fmt.Errorf("tcp connection failed: %w", err)
	}

	x509Certificates := conn.ConnectionState().PeerCertificates
	if len(x509Certificates) == 0 {
		return "", fmt.Errorf("no certificates returned from %s", addr)
	}
	// get last certificate in the chain
	fingerprint := sha1.Sum(x509Certificates[len(x509Certificates)-1].Raw)
	return fmt.Sprintf("%x", fingerprint), nil
}

func getHostAndPort(addr string) (string, error) {
	u, err := url.Parse(addr)
	if err != nil {
		return "", err
	}
	if u.Scheme == "" {
		return getHostAndPort(fmt.Sprintf("https://%s", addr))
	}

	port := u.Port()
	if port != "0" && port != "" {
		return u.Host, nil
	}
	if u.Scheme == "http" {
		return fmt.Sprintf("%s:80", u.Host), nil
	}
	// default to 443
	return fmt.Sprintf("%s:443", u.Host), nil
}
