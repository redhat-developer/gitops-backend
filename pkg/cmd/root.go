package cmd

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"strings"

	argoV1aplha1 "github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/redhat-developer/gitops-backend/pkg/git"
	"github.com/redhat-developer/gitops-backend/pkg/health"
	"github.com/redhat-developer/gitops-backend/pkg/httpapi"
	"github.com/redhat-developer/gitops-backend/pkg/httpapi/secrets"
	"github.com/redhat-developer/gitops-backend/pkg/metrics"
)

const (
	portFlag            = "port"
	insecureFlag        = "insecure"
	tlsCertFlag         = "tls-cert"
	tlsKeyFlag          = "tls-key"
	noTLSFlag           = "no-tls"
	enableHTTP2         = "enable-http2"
	tlsMinVersionFlag   = "tlsminversion"
	tlsMaxVersionFlag   = "tlsmaxversion"
	tlsCipherSuitesFlag = "tlsciphersuites"
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

func init() {
	cobra.OnInitialize(initConfig)
	if err := argoV1aplha1.AddToScheme(scheme.Scheme); err != nil {
		log.Fatalf("failed to initialize ArgoCD scheme, err: %v", err)
	}
}

func logIfError(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func initConfig() {
	viper.AutomaticEnv()
}

func makeHTTPCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gitops-backend",
		Short: "provide a simple API for fetching information",
		RunE: func(cmd *cobra.Command, args []string) error {
			m := metrics.New("backend", nil)

			http.Handle("/metrics", promhttp.Handler())
			http.HandleFunc("/health", health.Handler)

			router, err := makeAPIRouter(m)
			if err != nil {
				return err
			}
			http.Handle("/", httpapi.AuthenticationMiddleware(router))

			listen := fmt.Sprintf(":%d", viper.GetInt(portFlag))
			log.Printf("listening on %s", listen)

			server := &http.Server{
				Addr: listen,
			}
			// Disable HTTP/2 to mitigate CVE-2023-39325 & CVE-2023-44487
			if !viper.GetBool(enableHTTP2) {
				log.Printf("Disabled HTTP/2 protocol")
				server.TLSNextProto = map[string]func(*http.Server, *tls.Conn, http.Handler){}
			}

			if viper.GetBool(noTLSFlag) {
				log.Println("TLS connections disabled")
				return server.ListenAndServe()
			}
			tlsConfig, err := buildTLSConfig()
			if err != nil {
				return err
			}
			server.TLSConfig = tlsConfig
			log.Printf("Using TLS from %q and %q", viper.GetString(tlsCertFlag), viper.GetString(tlsKeyFlag))
			return server.ListenAndServeTLS(viper.GetString(tlsCertFlag), viper.GetString(tlsKeyFlag))
		},
	}

	cmd.Flags().Int(
		portFlag,
		8080,
		"port to serve requests on",
	)
	logIfError(viper.BindPFlag(portFlag, cmd.Flags().Lookup(portFlag)))

	cmd.Flags().Bool(
		insecureFlag,
		false,
		"allow insecure TLS requests",
	)
	logIfError(viper.BindPFlag(insecureFlag, cmd.Flags().Lookup(insecureFlag)))

	cmd.Flags().String(
		tlsKeyFlag,
		"/etc/gitops/ssl/tls.key",
		"filename for the TLS key",
	)
	logIfError(viper.BindPFlag(tlsKeyFlag, cmd.Flags().Lookup(tlsKeyFlag)))

	cmd.Flags().String(
		tlsCertFlag,
		"/etc/gitops/ssl/tls.crt",
		"filename for the TLS certficate",
	)
	logIfError(viper.BindPFlag(tlsCertFlag, cmd.Flags().Lookup(tlsCertFlag)))

	cmd.Flags().Bool(
		noTLSFlag,
		false,
		"do not attempt to read TLS certificates",
	)
	logIfError(viper.BindPFlag(noTLSFlag, cmd.Flags().Lookup(noTLSFlag)))

	cmd.Flags().Bool(
		enableHTTP2,
		false,
		"enable HTTP/2 for the server",
	)
	logIfError(viper.BindPFlag(enableHTTP2, cmd.Flags().Lookup(enableHTTP2)))
	cmd.Flags().String(
		tlsMinVersionFlag,
		"",
		"minimum supported TLS version (1.2, 1.3)",
	)
	logIfError(viper.BindPFlag(tlsMinVersionFlag, cmd.Flags().Lookup(tlsMinVersionFlag)))

	cmd.Flags().String(
		tlsMaxVersionFlag,
		"",
		"maximum supported TLS version (1.2, 1.3)",
	)
	logIfError(viper.BindPFlag(tlsMaxVersionFlag, cmd.Flags().Lookup(tlsMaxVersionFlag)))

	cmd.Flags().String(
		tlsCipherSuitesFlag,
		"",
		"comma-separated list of TLS cipher suites",
	)
	logIfError(viper.BindPFlag(tlsCipherSuitesFlag, cmd.Flags().Lookup(tlsCipherSuitesFlag)))
	return cmd
}

// Execute is the main entry point into this component.
func Execute() {
	if err := makeHTTPCmd().Execute(); err != nil {
		log.Fatal(err)
	}
}

func makeClusterConfig() (*rest.Config, error) {
	clusterConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create a cluster config: %s", err)
	}
	return clusterConfig, nil
}

func makeAPIRouter(m metrics.Interface) (*httpapi.APIRouter, error) {
	config, err := makeClusterConfig()
	if err != nil {
		return nil, err
	}
	cf := git.NewClientFactory(m)
	secretGetter := secrets.NewFromConfig(
		&rest.Config{Host: config.Host},
		viper.GetBool(insecureFlag))
	k8sClient, err := ctrlclient.New(config, ctrlclient.Options{})
	if err != nil {
		return nil, err
	}
	router := httpapi.NewRouter(cf, secretGetter, k8sClient)
	return router, nil
}
