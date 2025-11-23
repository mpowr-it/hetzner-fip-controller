package fipcontroller

import (
	"context"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
)

func (controller *Controller) leaseLock(id string) (lock *resourcelock.LeaseLock) {
	lock = &resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name:      controller.Configuration.LeaseName,
			Namespace: controller.Configuration.Namespace,
		},
		Client: controller.KubernetesClient.CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: id,
		},
	}
	return
}

func (controller *Controller) leaderElectionConfig() (config leaderelection.LeaderElectionConfig) {
	config = leaderelection.LeaderElectionConfig{
		Lock:            controller.leaseLock(controller.Configuration.PodName),
		ReleaseOnCancel: true,
		LeaseDuration:   time.Duration(controller.Configuration.LeaseDuration) * time.Second,
		RenewDeadline:   time.Duration(controller.Configuration.LeaseRenewDeadline) * time.Second,
		RetryPeriod:     2 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: controller.onStartedLeading,
			OnStoppedLeading: controller.onStoppedLeading,
		},
	}
	return
}

// RunWithLeaderElection starts a leaderelection and will run the main logic when it becomes the leader
func (controller *Controller) RunWithLeaderElection(ctx context.Context) {
	leaderelection.RunOrDie(ctx, controller.leaderElectionConfig())

	// because the context is closed, the client should report errors
	_, err := controller.KubernetesClient.CoordinationV1().Leases(controller.Configuration.Namespace).Get(ctx, controller.Configuration.LeaseName, metav1.GetOptions{})
	if err == nil || !strings.Contains(err.Error(), "the leader is shutting down") {
		controller.Logger.Fatalf("Expected to get an error when trying to make a client call: %v", err)
	}
}

// onStartedLeading is called once this instance becomes the leader.
// Any unexpected error from Run() should be logged, not fatal, in order
// to avoid controller CrashLoopBackoff on transient API or network issues.
func (controller *Controller) onStartedLeading(ctx context.Context) {
	controller.Logger.Info("Started leading")

	// Run the main reconciliation loop.
	// Run() should normally only return when the context is cancelled.
	if err := controller.Run(ctx); err != nil {
		// Do not call Fatalf here — that would kill the process and trigger a pod restart.
		// We only log the error and allow the manager lifecycle to continue gracefully.
		controller.Logger.WithError(err).Error("controller exited with an unexpected error")
	}
}

// onStoppedLeading is called when leadership is lost.
func (controller *Controller) onStoppedLeading() {
	controller.Logger.Info("Stopped leading")
}
