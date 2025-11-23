package fipcontroller

import (
	"context"
	"fmt"
	"os"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"

	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"

	"github.com/mpowr/hetzner-fip-controller/internal/pkg/configuration"
)

// Controller is the main struct used for all other functions in this package.
// Holds all client configurations and loggers
type Controller struct {
	HetznerClient    *hcloud.Client
	KubernetesClient kubernetes.Interface
	Configuration    *configuration.Configuration
	Logger           *logrus.Logger
	Backoff          wait.Backoff
}

// NewController creates a new Controller and with it the client configurations and loggers
func NewController(config *configuration.Configuration) (*Controller, error) {
	// Validate controller config
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("controller config invalid: %v", err)
	}

	hetznerClient, err := newHetznerClient(config.HcloudAPIToken)
	if err != nil {
		return nil, fmt.Errorf("could not initialise hetzner client: %v", err)
	}

	kubernetesClient, err := newKubernetesClient()
	if err != nil {
		return nil, fmt.Errorf("could not initialise kubernetes client: %v", err)
	}

	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})
	logger.SetReportCaller(true)
	logger.SetOutput(os.Stdout)

	loglevel, err := logrus.ParseLevel(config.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("could not parse log level: %v", err)
	}
	logger.SetLevel(loglevel)

	backoff := wait.Backoff{
		Duration: config.BackoffDuration,
		Factor:   config.BackoffFactor,
		Steps:    config.BackoffSteps,
	}

	return &Controller{
		HetznerClient:    hetznerClient,
		KubernetesClient: kubernetesClient,
		Configuration:    config,
		Logger:           logger,
		Backoff:          backoff,
	}, nil
}

// Run executes the reconciliation loop once this controller instance
// becomes the elected leader. The loop must never kill the process due to
// transient runtime errors (e.g. Hetzner API 503, network issues).
// --
// Returning errors here would bubble up into the leader election logic,
// causing Fatalf() → os.Exit(1) → CrashLoopBackoff. Therefore, errors
// are logged but the loop continues running.
func (controller *Controller) Run(ctx context.Context) error {

	// Initial reconciliation attempt.
	// Any failure here is logged but not treated as fatal.
	if err := controller.UpdateFloatingIPs(ctx); err != nil {
		controller.Logger.WithError(err).Error("initial reconciliation failed")
	} else {
		controller.Logger.Info("Initialization complete. Starting reconciliation")
	}

	// Execute periodic reconciliation.
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {

		// Normal shutdown path when the leader election context is cancelled.
		case <-ctx.Done():
			controller.Logger.Info("Context canceled, shutting down controller")
			return nil

		// Main reconcile interval.
		case <-ticker.C:
			// Never return an error from here — it would cause a fatal exit.
			if err := controller.UpdateFloatingIPs(ctx); err != nil {
				controller.Logger.WithError(err).Error("reconciliation failed")
			}
		}
	}
}

// UpdateFloatingIPs searches for running hetzner cloud servers and sort them by fewest assigned floating ips.
// It then (re)assigns all unassigned ips or ips that are assigned to non running servers to the sorted running serves.
func (controller *Controller) UpdateFloatingIPs(ctx context.Context) error {
	controller.Logger.Debugf("Checking floating IPs")

	// Get running servers for floating ip assignment
	nodeAddressList, err := controller.nodeAddressList(ctx, controller.Configuration.NodeAddressType)
	if err != nil {
		return fmt.Errorf("could not get addressList for active kubernetes nodes: %v", err)
	}

	if nodeAddressList == nil || len(nodeAddressList) < 1 {
		return fmt.Errorf("could not find any ips")
	}

	runningServers, err := controller.servers(ctx, nodeAddressList)
	if err != nil {
		return fmt.Errorf("could not get server objects for addressList: %v", err)
	}

	if runningServers == nil || len(runningServers) < 1 {
		return fmt.Errorf("no server objects were found")
	}

	// Get floatingIPs from config if specified, otherwise from hetzner api
	floatingIPs, err := controller.getFloatingIPs(ctx)
	if err != nil {
		return fmt.Errorf("could not get floatingIPs: %v", err)
	}

	for _, floatingIP := range floatingIPs {
		controller.Logger.Debugf("Checking floating IP: %s", floatingIP.IP.String())

		// (Re)assign floatingIP if no server is assigned or the assigned server is not running
		// Since we already have all running server in a slice we can just search through it
		if floatingIP.Server == nil || !hasServerByID(runningServers, floatingIP.Server) {
			// Get the server with the lowest amount of fips (cant be nil since we know that servers can't be empty)
			server := findServerWithLowestFIP(runningServers)

			controller.Logger.Infof("Switching address '%s' to server '%s'", floatingIP.IP.String(), server.Name)
			var response *hcloud.Response
			err = retry.OnError(controller.Backoff, alwaysRetry, func() error {
				_, response, err = controller.HetznerClient.FloatingIP.Assign(ctx, floatingIP, server)
				return err
			})
			if err != nil {
				return fmt.Errorf("could not update floating IP '%s': %v", floatingIP.IP.String(), err)
			}
			if response.StatusCode != 201 {
				return fmt.Errorf("could not update floating IP '%s': Got HTTP Code %d, expected 201", floatingIP.IP.String(), response.StatusCode)
			}
			// Add placeholder floating ip to server so that findServerWithLowestFIP will always get a correct server
			server.PublicNet.FloatingIPs = append(server.PublicNet.FloatingIPs, &hcloud.FloatingIP{})
		}

	}
	return nil
}

// Find the server with the lowest amount of floating IPs
func findServerWithLowestFIP(servers []*hcloud.Server) *hcloud.Server {
	if len(servers) < 1 {
		return nil
	}
	server := servers[0]
	for _, s := range servers {
		if len(s.PublicNet.FloatingIPs) < len(server.PublicNet.FloatingIPs) {
			server = s
		}
	}
	return server
}

// Checks for a server in a slice by its id
// Returns true the server was found
func hasServerByID(slice []*hcloud.Server, val *hcloud.Server) bool {
	for _, item := range slice {
		if item.ID == val.ID {
			return true
		}
	}
	return false
}

func alwaysRetry(_ error) bool {
	return true
}
