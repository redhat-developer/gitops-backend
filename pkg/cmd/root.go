package cmd

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"

	argoV1aplha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
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
	portFlag     = "port"
	insecureFlag = "insecure"
	tlsCertFlag  = "tls-cert"
	tlsKeyFlag   = "tls-key"
	noTLSFlag    = "no-tls"
	enableHTTP2  = "enable-http2"
)

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
