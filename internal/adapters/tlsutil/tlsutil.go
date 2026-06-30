package tlsutil

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strings"
)

// TLSConfigFromCAFile returns a TLS client config using the system trust store plus
// certificates loaded from caFile. An empty caFile is invalid here; callers should
// only call this function when a CA file path is explicitly configured.
func TLSConfigFromCAFile(caFile string) (*tls.Config, error) {
	caFile = strings.TrimSpace(caFile)
	if caFile == "" {
		return nil, fmt.Errorf("CA file path is required")
	}

	certPool, err := x509.SystemCertPool()
	if err != nil || certPool == nil {
		certPool = x509.NewCertPool()
	}

	pemData, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("read CA file %q: %w", caFile, err)
	}
	if ok := certPool.AppendCertsFromPEM(pemData); !ok {
		return nil, fmt.Errorf("CA file %q does not contain valid PEM certificates", caFile)
	}

	return &tls.Config{RootCAs: certPool, MinVersion: tls.VersionTLS12}, nil
}
