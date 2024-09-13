package cmd

import (
	"fmt"

	"github.com/kubernetes-sigs/ingress2gateway/pkg/i2gw"
	"github.com/spf13/cobra"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

var (
	scheme               = runtime.NewScheme()
	setupLog             = ctrl.Log.WithName("setup")
	enableLeaderElection bool
	probeAddr            string
	providers            []string
	gateway              string
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(networkingv1.AddToScheme(scheme))
	utilruntime.Must(gatewayv1.Install(scheme))
}

func runController(cmd *cobra.Command, _ []string) error {
	config := ctrl.GetConfigOrDie()
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	mgrOptions := ctrl.Options{
		Scheme:                 scheme,
		HealthProbeBindAddress: probeAddr,
		Metrics: server.Options{
			BindAddress: ":9443",
		},
		LeaderElection:   enableLeaderElection,
		LeaderElectionID: "79cbe7f4.networking.k8s.io",
	}

	mgr, err := ctrl.NewManager(config, mgrOptions)
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		return err
	}

	providerByName, err := i2gw.ConstructProviders(&i2gw.ProviderConf{
		Client:  mgr.GetClient(),
		Gateway: gateway,
		Scheme:  mgr.GetScheme(),
	}, providers)
	if err != nil {
		return err
	}

	for _, provider := range providerByName {
		if err = provider.SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "Ingress")
			return err
		}
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		return err
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		return err
	}

	setupLog.Info("starting manager")

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		return err
	}

	return nil
}

func newControllerCommand() *cobra.Command {
	// controller represents the controller command. It runs controller that converts Ingress resources to HTTPRoutes and Gateways
	// and applies to kubernetes cluster.
	var cmd = &cobra.Command{
		Use:   "controller",
		Short: "Runs the controller that converts Ingress resources to HTTPRoutes and Gateways and applies to kubernetes cluster.",
		RunE:  runController,
	}
	cmd.Flags().BoolVar(&enableLeaderElection, "leader-elect", false, "Enable leader election for controller manager")
	cmd.Flags().StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	cmd.Flags().StringSliceVar(&providers, "providers", i2gw.GetSupportedProviders(),
		fmt.Sprintf("If present, the tool will try to convert only resources related to the specified providers, supported values are %v.", i2gw.GetSupportedProviders()))
	cmd.Flags().StringVarP(&gateway, "gateway", "g", "", `If present, set as parent for all httpRoutes.`)

	return cmd
}
