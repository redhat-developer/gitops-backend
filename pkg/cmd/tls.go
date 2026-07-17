package cmd

import (
	"crypto/tls"
	"fmt"
	"log"
	"strings"

	"github.com/spf13/viper"
)

const (
	tlsMinVersionFlag   = "tls-min-version"
	tlsMaxVersionFlag   = "tls-max-version"
	tlsCipherSuitesFlag = "tls-cipher-suites"
)

type TLSConfig struct {
	MinVersion string
	MaxVersion string
	Ciphers    string
}

// tlsVersionMap maps version strings to tls version constants.
// TLS 1.0 is not supported as it is considered insecure.
var tlsVersionMap = map[string]uint16{
	"1.1":    tls.VersionTLS11,
	"tls1.1": tls.VersionTLS11,
	"1.2":    tls.VersionTLS12,
	"tls1.2": tls.VersionTLS12,
	"1.3":    tls.VersionTLS13,
	"tls1.3": tls.VersionTLS13,
}

// TLSVersionName returns a human-readable name for a TLS version constant.
func TLSVersionName(version uint16) string {
	for name, v := range tlsVersionMap {
		if v == version {
			return name
		}
	}
	return fmt.Sprintf("unknown (%d)", version)
}

// ParseTLSVersion parses a TLS version string (e.g. "1.2", "1.3", "TLS1.2") into a tls version constant.
// Returns 0 if the string is empty (meaning "use default").
func ParseTLSVersion(version string) (uint16, error) {
	if version == "" {
		return 0, nil
	}
	v, ok := tlsVersionMap[strings.ToLower(strings.TrimSpace(version))]
	if !ok {
		return 0, fmt.Errorf("unsupported TLS version: %q (supported: 1.1, 1.2, 1.3)", version)
	}
	return v, nil
}

// ParseTLSCiphers parses a colon-separated list of cipher suite names into cipher suite IDs.
// Only secure cipher suites (from tls.CipherSuites()) are allowed.
// Returns nil if the input is empty.
func ParseTLSCiphers(ciphers string) ([]uint16, error) {
	if ciphers == "" {
		return nil, nil
	}
	// Build lookup map from Go's secure cipher suites only
	cipherMap := make(map[string]uint16)
	for _, cs := range tls.CipherSuites() {
		cipherMap[cs.Name] = cs.ID
	}
	var result []uint16
	for _, name := range strings.Split(ciphers, ":") {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		id, ok := cipherMap[name]
		if !ok {
			return nil, fmt.Errorf("unsupported TLS cipher suite: %q", name)
		}
		result = append(result, id)
	}
	return result, nil
}

// ValidateTLSConfig validates the TLS configuration parameters.
// It checks that:
//   - The minimum TLS version is not greater than the maximum TLS version
//   - All configured cipher suites are compatible with the minimum TLS version
func ValidateTLSConfig(minVersion, maxVersion uint16, cipherSuites []uint16) error {
	if minVersion != 0 && maxVersion != 0 && minVersion > maxVersion {
		return fmt.Errorf("minimum TLS version (%s) cannot be higher than maximum TLS version (%s)",
			TLSVersionName(minVersion), TLSVersionName(maxVersion))
	}
	if len(cipherSuites) > 0 && minVersion != 0 {
		availableCiphers := tls.CipherSuites()
		for _, cipherID := range cipherSuites {
			for _, cs := range availableCiphers {
				if cs.ID == cipherID {
					supported := false
					for _, v := range cs.SupportedVersions {
						if v == minVersion {
							supported = true
							break
						}
					}
					if !supported {
						return fmt.Errorf("cipher suite %s is not supported by minimum TLS version %s",
							cs.Name, TLSVersionName(minVersion))
					}
					break
				}
			}
		}
	}

	return nil
}

func buildTLSConfig() (*tls.Config, error) {
	cfg := TLSConfig{
		MinVersion: viper.GetString(tlsMinVersionFlag),
		MaxVersion: viper.GetString(tlsMaxVersionFlag),
		Ciphers:    viper.GetString(tlsCipherSuitesFlag),
	}

	tlsCfg := &tls.Config{} //nolint:gosec
	minVer, err := ParseTLSVersion(cfg.MinVersion)
	if err != nil {
		return nil, fmt.Errorf("invalid --%s: %w", tlsMinVersionFlag, err)
	}
	tlsCfg.MinVersion = minVer
	maxVer, err := ParseTLSVersion(cfg.MaxVersion)
	if err != nil {
		return nil, fmt.Errorf("invalid --%s: %w", tlsMaxVersionFlag, err)
	}
	tlsCfg.MaxVersion = maxVer
	ciphers, err := ParseTLSCiphers(cfg.Ciphers)
	if err != nil {
		return nil, fmt.Errorf("invalid --%s: %w", tlsCipherSuitesFlag, err)
	}
	if len(ciphers) > 0 && minVer == tls.VersionTLS13 {
		log.Printf(
			"%s has no effect when minimum TLS version is 1.3 because TLS 1.3 cipher suites are not configurable; ignoring",
			tlsCipherSuitesFlag,
		)
		ciphers = nil
	}
	if err := ValidateTLSConfig(minVer, maxVer, ciphers); err != nil {
		return nil, err
	}
	tlsCfg.CipherSuites = ciphers

	return tlsCfg, nil
}
