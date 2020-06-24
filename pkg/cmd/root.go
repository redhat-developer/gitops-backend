package cmd

import (
	"fmt"
	"log"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rhd-gitops-examples/gitops-backend/pkg/httpapi"
	"github.com/rhd-gitops-examples/gitops-backend/pkg/httpapi/secrets"
	"github.com/rhd-gitops-examples/gitops-backend/pkg/metrics"
)

const (
	portFlag = "port"
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
			listen := fmt.Sprintf(":%d", viper.GetInt(portFlag))
			log.Printf("listening on %s", listen)
			http.Handle("/metrics", promhttp.Handler())
			client, err := makeClient()
			if err != nil {
				return err
			}

			cf := httpapi.NewClientFactory(httpapi.NewDriverIdentifier(), m)
			router := httpapi.NewRouter(cf, secrets.New(client))
			router.SecretRef = types.NamespacedName{
				Name:      "gitops-backend-secret",
				Namespace: "pipelines-app-delivery",
			}
			http.Handle("/", router)
			return http.ListenAndServe(listen, nil)
		},
	}

	cmd.Flags().Int(
		portFlag,
		8080,
		"port to serve requests on",
	)
	logIfError(viper.BindPFlag(portFlag, cmd.Flags().Lookup(portFlag)))
	return cmd
}

// Execute is the main entry point into this component.
func Execute() {
	if err := makeHTTPCmd().Execute(); err != nil {
		log.Fatal(err)
	}
}

func makeClient() (kubernetes.Interface, error) {
	clusterConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create a cluster config: %s", err)
	}
	c, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create the core client: %v", err)
	}
	return c, err
}
