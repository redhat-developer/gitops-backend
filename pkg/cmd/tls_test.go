package cmd

import (
	"crypto/tls"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTLSVersionName(t *testing.T) {
	tests := []struct {
		name     string
		version  uint16
		expected string
	}{
		{
			name:     "TLS1.1",
			version:  tls.VersionTLS11,
			expected: "1.1",
		},
		{
			name:     "TLS1.2",
			version:  tls.VersionTLS12,
			expected: "1.2",
		},
		{
			name:     "TLS1.3",
			version:  tls.VersionTLS13,
			expected: "1.3",
		},
		{
			name:     "unknown",
			version:  999,
			expected: "unknown (999)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, TLSVersionName(tt.version))
		})
	}
}

func TestParseTLSVersion(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  uint16
		expectErr bool
	}{
		{
			name:     "empty",
			input:    "",
			expected: 0,
		},
		{
			name:     "1.2",
			input:    "1.2",
			expected: tls.VersionTLS12,
		},
		{
			name:     "TLS1.2",
			input:    "TLS1.2",
			expected: tls.VersionTLS12,
		},
		{
			name:     "trim spaces",
			input:    " 1.3 ",
			expected: tls.VersionTLS13,
		},
		{
			name:      "unsupported",
			input:     "1.0",
			expectErr: true,
		},
		{
			name:      "garbage",
			input:     "abc",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, err := ParseTLSVersion(tt.input)

			if tt.expectErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, version)
		})
	}
}

func TestParseTLSCiphers(t *testing.T) {
	validCipher := tls.CipherSuites()[0]

	tests := []struct {
		name      string
		input     string
		expected  []uint16
		expectErr bool
	}{
		{
			name:     "empty",
			input:    "",
			expected: nil,
		},
		{
			name:     "single cipher",
			input:    validCipher.Name,
			expected: []uint16{validCipher.ID},
		},
		{
			name: "multiple ciphers",
			input: validCipher.Name + ":" +
				tls.CipherSuites()[1].Name,
			expected: []uint16{
				validCipher.ID,
				tls.CipherSuites()[1].ID,
			},
		},
		{
			name:      "invalid cipher",
			input:     "INVALID_CIPHER",
			expectErr: true,
		},
		{
			name:     "ignore empty entries",
			input:    ":" + validCipher.Name + "::",
			expected: []uint16{validCipher.ID},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ciphers, err := ParseTLSCiphers(tt.input)

			if tt.expectErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, ciphers)
		})
	}
}

func TestValidateTLSConfig(t *testing.T) {
	tls12cipher := tls.CipherSuites()[3]

	tests := []struct {
		name       string
		minVersion uint16
		maxVersion uint16
		ciphers    []uint16
		expectErr  bool
	}{
		{
			name:       "valid versions",
			minVersion: tls.VersionTLS12,
			maxVersion: tls.VersionTLS13,
		},
		{
			name:       "min greater than max",
			minVersion: tls.VersionTLS13,
			maxVersion: tls.VersionTLS12,
			expectErr:  true,
		},
		{
			name:       "no versions",
			minVersion: 0,
			maxVersion: 0,
		},
		{
			name:       "valid cipher",
			minVersion: tls.VersionTLS12,
			maxVersion: tls.VersionTLS13,
			ciphers:    []uint16{tls12cipher.ID},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTLSConfig(
				tt.minVersion,
				tt.maxVersion,
				tt.ciphers,
			)

			if tt.expectErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestParseTLSCurvePreferences(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  []tls.CurveID
		expectErr bool
	}{
		{
			name:     "empty",
			input:    "",
			expected: nil,
		},
		{
			name:     "single curve",
			input:    tls.X25519.String(),
			expected: []tls.CurveID{tls.X25519},
		},
		{
			name: "multiple curves",
			input: tls.X25519MLKEM768.String() + ":" +
				tls.X25519.String() + ":" +
				tls.CurveP256.String(),
			expected: []tls.CurveID{
				tls.X25519MLKEM768,
				tls.X25519,
				tls.CurveP256,
			},
		},
		{
			name: "post-quantum curves",
			input: tls.SecP256r1MLKEM768.String() + ":" +
				tls.SecP384r1MLKEM1024.String(),
			expected: []tls.CurveID{
				tls.SecP256r1MLKEM768,
				tls.SecP384r1MLKEM1024,
			},
		},
		{
			name:      "invalid curve",
			input:     "INVALID_CURVE",
			expectErr: true,
		},
		{
			name:     "ignore empty entries",
			input:    ":" + tls.X25519.String() + "::",
			expected: []tls.CurveID{tls.X25519},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			curves, err := ParseTLSCurvePreferences(tt.input)

			if tt.expectErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, curves)
		})
	}
}

func TestBuildTLSConfig(t *testing.T) {
	validCipher := tls.CipherSuites()[3]

	tests := []struct {
		name           string
		minVersion     string
		maxVersion     string
		ciphers        string
		curves         string
		expectedMin    uint16
		expectedMax    uint16
		expectedSuite  []uint16
		expectedCurves []tls.CurveID
		expectErr      bool
	}{
		{
			name:        "defaults",
			expectedMin: 0,
			expectedMax: 0,
		},
		{
			name:        "valid config",
			minVersion:  "1.2",
			maxVersion:  "1.3",
			ciphers:     validCipher.Name,
			expectedMin: tls.VersionTLS12,
			expectedMax: tls.VersionTLS13,
			expectedSuite: []uint16{
				validCipher.ID,
			},
		},
		{
			name:           "valid curves",
			curves:         tls.X25519.String() + ":" + tls.CurveP256.String(),
			expectedCurves: []tls.CurveID{tls.X25519, tls.CurveP256},
		},
		{
			name:       "invalid min version",
			minVersion: "1.0",
			expectErr:  true,
		},
		{
			name:       "invalid max version",
			maxVersion: "abc",
			expectErr:  true,
		},
		{
			name:      "invalid cipher",
			ciphers:   "BAD_CIPHER",
			expectErr: true,
		},
		{
			name:      "invalid curve",
			curves:    "BAD_CURVE",
			expectErr: true,
		},
		{
			name:       "min greater than max",
			minVersion: "1.3",
			maxVersion: "1.2",
			expectErr:  true,
		},
		{
			name:        "tls13 ignores ciphers",
			minVersion:  "1.3",
			ciphers:     validCipher.Name,
			expectedMin: tls.VersionTLS13,
			expectedMax: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()

			viper.Set(tlsMinVersionFlag, tt.minVersion)
			viper.Set(tlsMaxVersionFlag, tt.maxVersion)
			viper.Set(tlsCipherSuitesFlag, tt.ciphers)
			viper.Set(tlsCurvePreferencesFlag, tt.curves)

			cfg, err := buildTLSConfig()

			if tt.expectErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedMin, cfg.MinVersion)
			assert.Equal(t, tt.expectedMax, cfg.MaxVersion)
			assert.Equal(t, tt.expectedSuite, cfg.CipherSuites)
			assert.Equal(t, tt.expectedCurves, cfg.CurvePreferences)
		})
	}
}
