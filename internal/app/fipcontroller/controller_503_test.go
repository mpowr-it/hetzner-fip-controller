package fipcontroller

import (
	"context"
	"encoding/json"
	"net/http"
	"sync/atomic"
	"testing"

	"github.com/hetznercloud/hcloud-go/hcloud/schema"
	"github.com/mpowr/hetzner-fip-controller/internal/pkg/configuration"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes/fake"
)

// Test503Resilience verifies the controller continues operating despite 503 errors
func Test503Resilience(t *testing.T) {
	tests := []struct {
		name                  string
		simulateNodeList503   bool
		simulateServer503     bool
		simulateFloatingIP503 bool
		simulateAssign503     bool
		expectError           bool
		expectAssignmentCall  bool
	}{
		{
			name:                  "all APIs return 503 - should not error",
			simulateNodeList503:   true,
			simulateServer503:     true,
			simulateFloatingIP503: true,
			simulateAssign503:     true,
			expectError:           false,
			expectAssignmentCall:  false,
		},
		{
			name:                  "only floating IP fetch returns 503",
			simulateNodeList503:   false,
			simulateServer503:     false,
			simulateFloatingIP503: true,
			simulateAssign503:     false,
			expectError:           false,
			expectAssignmentCall:  false,
		},
		{
			name:                  "only server fetch returns 503",
			simulateNodeList503:   false,
			simulateServer503:     true,
			simulateFloatingIP503: false,
			simulateAssign503:     false,
			expectError:           false,
			expectAssignmentCall:  false,
		},
		{
			name:                  "only assignment returns 503",
			simulateNodeList503:   false,
			simulateServer503:     false,
			simulateFloatingIP503: false,
			simulateAssign503:     true,
			expectError:           false,
			expectAssignmentCall:  true, // Assignment is called but fails with 503
		},
		{
			name:                  "no 503 errors - successful assignment",
			simulateNodeList503:   false,
			simulateServer503:     false,
			simulateFloatingIP503: false,
			simulateAssign503:     false,
			expectError:           false,
			expectAssignmentCall:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Setup test environment
			testEnv := newTestEnv()
			defer testEnv.Teardown()

			// Track API calls
			var assignCalled atomic.Bool
			var serversCalled atomic.Int32
			var floatingIPsCalled atomic.Int32

			// Configure mock Hetzner API endpoints
			testEnv.Mux.HandleFunc("/servers", func(w http.ResponseWriter, r *http.Request) {
				serversCalled.Add(1)
				if test.simulateServer503 && serversCalled.Load() <= 5 {
					w.WriteHeader(http.StatusServiceUnavailable)
					w.Write([]byte("Service Unavailable"))
					return
				}

				json.NewEncoder(w).Encode(schema.ServerListResponse{
					Servers: []schema.Server{
						{
							ID:   1,
							Name: "server-1",
							PublicNet: schema.ServerPublicNet{
								IPv4: schema.ServerPublicNetIPv4{
									IP: "1.2.3.4",
								},
							},
						},
					},
				})
			})

			testEnv.Mux.HandleFunc("/floating_ips", func(w http.ResponseWriter, r *http.Request) {
				if r.Method == "GET" {
					floatingIPsCalled.Add(1)
					if test.simulateFloatingIP503 && floatingIPsCalled.Load() <= 5 {
						w.WriteHeader(http.StatusServiceUnavailable)
						w.Write([]byte("Service Unavailable"))
						return
					}

					json.NewEncoder(w).Encode(schema.FloatingIPListResponse{
						FloatingIPs: []schema.FloatingIP{
							{
								ID:     1,
								Type:   "ipv4",
								IP:     "1.2.3.4",
								Server: nil, // Not assigned
							},
						},
					})
				}
			})

			testEnv.Mux.HandleFunc("/floating_ips/1/actions/assign", func(w http.ResponseWriter, r *http.Request) {
				assignCalled.Store(true)
				if test.simulateAssign503 {
					w.WriteHeader(http.StatusServiceUnavailable)
					w.Write([]byte("Service Unavailable"))
					return
				}

				w.WriteHeader(http.StatusCreated)
				json.NewEncoder(w).Encode(schema.FloatingIPActionAssignResponse{
					Action: schema.Action{
						ID: 1,
					},
				})
			})

			// Create Kubernetes client with test node
			kubeClient := fake.NewSimpleClientset(
				createTestNode("node-1", []v1.NodeAddress{
					{
						Type:    v1.NodeExternalIP,
						Address: "1.2.3.4",
					},
				}, v1.ConditionTrue),
			)

			// Setup controller with retry configuration
			controller := Controller{
				HetznerClient:    testEnv.Client,
				KubernetesClient: kubeClient,
				Backoff: wait.Backoff{
					Steps:    5,
					Duration: 1,
					Factor:   1.2,
				},
				Configuration: &configuration.Configuration{
					NodeAddressType: configuration.NodeAddressTypeExternal,
				},
				Logger: logrus.New(),
			}

			// Silence logs during test
			controller.Logger.SetLevel(logrus.PanicLevel)

			// Execute UpdateFloatingIPs
			err := controller.UpdateFloatingIPs(context.Background())

			// Verify expectations
			if test.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !test.expectError && err != nil {
				t.Errorf("expected no error but got: %v", err)
			}
			if test.expectAssignmentCall && !assignCalled.Load() {
				t.Errorf("expected assignment to be called but it wasn't")
			}
			if !test.expectAssignmentCall && assignCalled.Load() {
				t.Errorf("expected assignment not to be called but it was")
			}
		})
	}
}

// TestPartial503Recovery tests that the controller recovers when some APIs succeed after retries
func TestPartial503Recovery(t *testing.T) {
	testEnv := newTestEnv()
	defer testEnv.Teardown()

	// Track call counts
	var serverCallCount atomic.Int32
	var floatingIPCallCount atomic.Int32
	var assignCallCount atomic.Int32

	// Fail first 3 attempts, then succeed
	testEnv.Mux.HandleFunc("/servers", func(w http.ResponseWriter, r *http.Request) {
		count := serverCallCount.Add(1)
		if count <= 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		json.NewEncoder(w).Encode(schema.ServerListResponse{
			Servers: []schema.Server{
				{
					ID:   1,
					Name: "server-1",
					PublicNet: schema.ServerPublicNet{
						IPv4: schema.ServerPublicNetIPv4{
							IP: "1.2.3.4",
						},
					},
				},
			},
		})
	})

	testEnv.Mux.HandleFunc("/floating_ips", func(w http.ResponseWriter, r *http.Request) {
		count := floatingIPCallCount.Add(1)
		if count <= 2 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		json.NewEncoder(w).Encode(schema.FloatingIPListResponse{
			FloatingIPs: []schema.FloatingIP{
				{
					ID:     1,
					Type:   "ipv4",
					IP:     "1.2.3.4",
					Server: nil,
				},
			},
		})
	})

	testEnv.Mux.HandleFunc("/floating_ips/1/actions/assign", func(w http.ResponseWriter, r *http.Request) {
		count := assignCallCount.Add(1)
		if count <= 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(schema.FloatingIPActionAssignResponse{
			Action: schema.Action{ID: 1},
		})
	})

	// Setup Kubernetes client
	kubeClient := fake.NewSimpleClientset(
		createTestNode("node-1", []v1.NodeAddress{
			{Type: v1.NodeExternalIP, Address: "1.2.3.4"},
		}, v1.ConditionTrue),
	)

	// Create controller
	controller := Controller{
		HetznerClient:    testEnv.Client,
		KubernetesClient: kubeClient,
		Backoff: wait.Backoff{
			Steps:    5,
			Duration: 1,
			Factor:   1.1,
		},
		Configuration: &configuration.Configuration{
			NodeAddressType: configuration.NodeAddressTypeExternal,
		},
		Logger: logrus.New(),
	}
	controller.Logger.SetLevel(logrus.PanicLevel)

	// Run update
	err := controller.UpdateFloatingIPs(context.Background())
	// Should succeed despite initial 503s
	if err != nil {
		t.Errorf("expected no error but got: %v", err)
	}

	// Verify retries occurred
	if serverCallCount.Load() <= 3 {
		t.Errorf("expected more than 3 server calls due to retries, got %d", serverCallCount.Load())
	}
	if floatingIPCallCount.Load() <= 2 {
		t.Errorf("expected more than 2 floating IP calls due to retries, got %d", floatingIPCallCount.Load())
	}
	if assignCallCount.Load() <= 1 {
		t.Errorf("expected more than 1 assign call due to retries, got %d", assignCallCount.Load())
	}
}

// TestMultipleFloatingIPsWith503 tests handling of multiple floating IPs with partial 503 errors
func TestMultipleFloatingIPsWith503(t *testing.T) {
	testEnv := newTestEnv()
	defer testEnv.Teardown()

	assignedIPs := make(map[int]bool)

	// Setup endpoints
	testEnv.Mux.HandleFunc("/servers", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(schema.ServerListResponse{
			Servers: []schema.Server{
				{
					ID:   1,
					Name: "server-1",
					PublicNet: schema.ServerPublicNet{
						IPv4: schema.ServerPublicNetIPv4{
							IP: "10.0.0.1",
						},
					},
				},
			},
		})
	})

	testEnv.Mux.HandleFunc("/floating_ips", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(schema.FloatingIPListResponse{
			FloatingIPs: []schema.FloatingIP{
				{ID: 1, Type: "ipv4", IP: "1.1.1.1", Server: nil},
				{ID: 2, Type: "ipv4", IP: "2.2.2.2", Server: nil},
				{ID: 3, Type: "ipv4", IP: "3.3.3.3", Server: nil},
			},
		})
	})

	// Fail assignment for IP 2, succeed for others
	testEnv.Mux.HandleFunc("/floating_ips/1/actions/assign", func(w http.ResponseWriter, r *http.Request) {
		assignedIPs[1] = true
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(schema.FloatingIPActionAssignResponse{Action: schema.Action{ID: 1}})
	})

	testEnv.Mux.HandleFunc("/floating_ips/2/actions/assign", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	})

	testEnv.Mux.HandleFunc("/floating_ips/3/actions/assign", func(w http.ResponseWriter, r *http.Request) {
		assignedIPs[3] = true
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(schema.FloatingIPActionAssignResponse{Action: schema.Action{ID: 3}})
	})

	// Setup Kubernetes client
	kubeClient := fake.NewSimpleClientset(
		createTestNode("node-1", []v1.NodeAddress{
			{Type: v1.NodeExternalIP, Address: "10.0.0.1"},
		}, v1.ConditionTrue),
	)

	// Create controller
	controller := Controller{
		HetznerClient:    testEnv.Client,
		KubernetesClient: kubeClient,
		Backoff: wait.Backoff{
			Steps:    3,
			Duration: 1,
			Factor:   1.1,
		},
		Configuration: &configuration.Configuration{
			NodeAddressType: configuration.NodeAddressTypeExternal,
		},
		Logger: logrus.New(),
	}
	controller.Logger.SetLevel(logrus.PanicLevel)

	// Run update
	err := controller.UpdateFloatingIPs(context.Background())
	// Should not error despite one IP failing
	if err != nil {
		t.Errorf("expected no error but got: %v", err)
	}

	// Verify IPs 1 and 3 were assigned, but not 2
	if !assignedIPs[1] {
		t.Error("expected IP 1 to be assigned")
	}
	if assignedIPs[2] {
		t.Error("expected IP 2 not to be assigned due to 503")
	}
	if !assignedIPs[3] {
		t.Error("expected IP 3 to be assigned")
	}
}
