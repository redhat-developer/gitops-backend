package cmd

import (
	"fmt"
	"log"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/client-go/rest"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rhd-gitops-examples/gitops-backend/pkg/git"
	"github.com/rhd-gitops-examples/gitops-backend/pkg/health"
	"github.com/rhd-gitops-examples/gitops-backend/pkg/httpapi"
	"github.com/rhd-gitops-examples/gitops-backend/pkg/httpapi/secrets"
	"github.com/rhd-gitops-examples/gitops-backend/pkg/metrics"
)

const (
	portFlag     = "port"
	insecureFlag = "insecure"
	tlsCertFlag  = "tls-cert"
	tlsKeyFlag   = "tls-key"
	noTLSFlag    = "no-tls"
)

func init() {
	cobra.OnInitialize(initConfig)
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
			if viper.GetBool(noTLSFlag) {
				log.Println("TLS connections disabled")
				return http.ListenAndServe(listen, nil)
			}
			log.Printf("Using TLS from %q and %q", viper.GetString(tlsCertFlag), viper.GetString(tlsKeyFlag))
			return http.ListenAndServeTLS(listen, viper.GetString(tlsCertFlag), viper.GetString(tlsKeyFlag), nil)
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
	cf := git.NewClientFactory(git.NewDriverIdentifier(), m)
	secretGetter := secrets.NewFromConfig(
		&rest.Config{Host: config.Host},
		viper.GetBool(insecureFlag))
	router := httpapi.NewRouter(cf, secretGetter)
	return router, nil
}
