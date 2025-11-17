# HETZNER FloatingIP Controller

[![Release](https://img.shields.io/github/v/release/mpowr-it/hetzner-fip-controller)](https://github.com/mpowr-it/hetzner-fip-controller/releases)
[![Docker Pulls](https://img.shields.io/docker/pulls/mpowr/hetzner-fip-controller)](https://hub.docker.com/r/mpowr/hetzner-fip-controller)
[![Go Report Card](https://goreportcard.com/badge/github.com/mpowr-it/hetzner-fip-controller)](https://goreportcard.com/report/github.com/mpowr-it/hetzner-fip-controller)
[![Security Scan](https://github.com/mpowr-it/hetzner-fip-controller/workflows/Release%20Docker%20Image/badge.svg)](https://github.com/mpowr-it/hetzner-fip-controller/actions)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/mpowr-it/hetzner-fip-controller/badge)](https://securityscorecards.dev/viewer/?uri=github.com/mpowr-it/hetzner-fip-controller)

> 🔧 **Maintained fork** of [`cbeneke/hcloud-fip-controller`](https://github.com/cbeneke/hcloud-fip-controller)

A robust Kubernetes controller for managing Hetzner Cloud Floating IPs (FIPs). Automatically assigns floating IPs to healthy nodes based on pod labels and node IPs.

## ✨ Features

- 🎯 **Leader Election** – Only one active controller instance
- 🔄 **Periodic Reconciliation** – Automatic floating IP reassignment
- 🏥 **Health & Readiness Probes** – Kubernetes-native health checks
- 🛡️ **Resilient** – Handles transient Hetzner API issues gracefully
- 📊 **Structured Logging** – Detailed, parseable logs
- 🔐 **Security Hardened** – Signed images, SBOMs, CVE scanning
- 📦 **Distroless Runtime** – Minimal attack surface (Chainguard Static)
- 🌍 **Multi-Architecture** – Supports `linux/amd64` and `linux/arm64`

## 🎯 Why This Fork?

The original project is no longer actively maintained. This fork addresses critical issues:

| Issue                  | Original                  | This Fork                        |
|------------------------|---------------------------|----------------------------------|
| **API Error Handling** | ❌ Crashes on `503` errors | ✅ Retry with exponential backoff |
| **Health Endpoints**   | ❌ None                    | ✅ `/healthz` and `/readyz`       |
| **Observability**      | ⚠️ Limited logging        | ✅ Structured, detailed logs      |
| **Container Security** | ⚠️ Alpine-based           | ✅ Distroless, signed, SBOM       |
| **Active Maintenance** | ❌ Archived                | ✅ Actively maintained            |

## 🚀 Quick Start

### Prerequisites

- Kubernetes cluster running on Hetzner Cloud
- Hetzner Cloud API token with read/write access
- `kubectl` configured to access your cluster

### Installation

#### 1. Create API Token Secret

```bash
kubectl create namespace fip-controller
kubectl create secret generic hcloud-token \
  --namespace fip-controller \
  --from-literal=token=YOUR_HETZNER_API_TOKEN
```

#### 2. Deploy the Controller

```bash
kubectl apply -f https://raw.githubusercontent.com/mpowr-it/hetzner-fip-controller/main/deploy/manifests.yaml
```

#### 3. Configure Floating IPs

Edit the ConfigMap to specify your floating IPs:

```bash
kubectl edit configmap fip-controller-config -n fip-controller
```

Add your floating IPs:

```yaml
data:
  HCLOUD_FLOATING_IPS: "1.2.3.4,5.6.7.8"
  NAMESPACE: "default"
  POD_LABEL_SELECTOR: "app=myapp"
```

#### 4. Verify Deployment

```bash
kubectl get pods -n fip-controller
kubectl logs -n fip-controller -l app=fip-controller
```

### Using Helm (Recommended)

```bash
helm repo add mpowr https://charts.mpowr.io
helm repo update
helm install hetzner-fip-controller mpowr/hetzner-fip-controller \
  --namespace fip-controller \
  --create-namespace \
  --set hcloudToken=YOUR_HETZNER_API_TOKEN \
  --set floatingIPs="{1.2.3.4,5.6.7.8}"
```

## 🔐 Security & Verification

### Image Verification

All container images are signed with [Sigstore/Cosign](https://sigstore.dev/):

```bash
# Verify image signature
cosign verify \
  --certificate-identity-regexp=https://github.com/mpowr-it/hetzner-fip-controller \
  --certificate-oidc-issuer=https://token.actions.githubusercontent.com \
  mpowr/hetzner-fip-controller:latest

# Download SBOM
cosign download sbom mpowr/hetzner-fip-controller:latest > sbom.json

# Verify attestations
cosign verify-attestation \
  --certificate-identity-regexp=https://github.com/mpowr-it/hetzner-fip-controller \
  --certificate-oidc-issuer=https://token.actions.githubusercontent.com \
  --type=https://spdx.dev/Document \
  mpowr/hetzner-fip-controller:latest
```

### Security Features

- ✅ **Signed Images** – Verifiable with Cosign
- ✅ **SBOM Included** – Software Bill of Materials (SPDX + CycloneDX)
- ✅ **CVE Scanning** – Automated with Trivy & Grype
- ✅ **Distroless Base** – Chainguard Static (zero packages)
- ✅ **Non-Root User** – Runs as UID 65532
- ✅ **Read-Only Root FS** – Immutable container filesystem
- ✅ **No Shell** – Attack surface minimization

### Vulnerability Reports

Check the [Security Tab](https://github.com/mpowr-it/hetzner-fip-controller/security) for the latest vulnerability scan results.

## 🏗️ Building from Source

### Local Build

```bash
# Build binary
make build

# Run tests
make test

# Build Docker image
make docker-build

# Security scan
make docker-security
```

### Development Environment

This project uses [Devbox](https://www.jetpack.io/devbox/) for reproducible dev environments:

```bash
# Install devbox
curl -fsSL https://get.jetpack.io/devbox | bash

# Enter dev shell
devbox shell

# All tools (Go, golangci-lint, trivy, etc.) are now available
make all
```

## 📖 Configuration

### Environment Variables

| Variable               | Description                                  | Default       | Required |
|------------------------|----------------------------------------------|---------------|----------|
| `HCLOUD_TOKEN`         | Hetzner Cloud API token                      | -             | ✅        |
| `HCLOUD_FLOATING_IPS`  | Comma-separated list of floating IPs         | -             | ✅        |
| `NAMESPACE`            | Kubernetes namespace to watch                | `default`     | ❌        |
| `POD_NAME`             | Specific pod name to watch                   | -             | ❌        |
| `POD_LABEL_SELECTOR`   | Label selector for pods                      | -             | ❌        |
| `NODE_NAME`            | Kubernetes node name                         | Auto-detected | ❌        |
| `NODE_ADDRESS_TYPE`    | Address type to use (`external`, `internal`) | `external`    | ❌        |
| `LOG_LEVEL`            | Log level (`debug`, `info`, `warn`, `error`) | `info`        | ❌        |
| `LEASE_DURATION`       | Leader election lease duration (seconds)     | `15`          | ❌        |
| `LEASE_RENEW_DEADLINE` | Leader election renew deadline (seconds)     | `10`          | ❌        |
| `BACKOFF_DURATION`     | Initial backoff duration                     | `1s`          | ❌        |
| `BACKOFF_FACTOR`       | Backoff multiplier                           | `1.2`         | ❌        |
| `BACKOFF_STEPS`        | Maximum backoff retries                      | `5`           | ❌        |

### Example ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: fip-controller-config
  namespace: fip-controller
data:
  HCLOUD_FLOATING_IPS: "1.2.3.4,5.6.7.8"
  NAMESPACE: "production"
  POD_LABEL_SELECTOR: "app=nginx,tier=frontend"
  NODE_ADDRESS_TYPE: "external"
  LOG_LEVEL: "info"
  LEASE_DURATION: "15"
  LEASE_RENEW_DEADLINE: "10"
  BACKOFF_DURATION: "2s"
  BACKOFF_FACTOR: "1.5"
  BACKOFF_STEPS: "5"
```

## 🔄 How It Works

1. **Leader Election**: Multiple controller replicas elect a leader using Kubernetes leases
2. **Pod Discovery**: Watches pods matching the configured label selector
3. **Node Selection**: Identifies healthy nodes running the target pods
4. **IP Assignment**: Assigns floating IPs to the selected node via Hetzner API
5. **Reconciliation**: Periodically verifies and corrects floating IP assignments
6. **Failover**: Automatically reassigns IPs when nodes become unhealthy

## 📊 Monitoring

### Health Endpoints

- **Liveness Probe**: `GET /healthz` – Returns `200` if controller is running
- **Readiness Probe**: `GET /readyz` – Returns `200` if controller is ready to handle requests

### Example Kubernetes Probes

```yaml
livenessProbe:
  httpGet:
    path: /healthz
    port: 8080
  initialDelaySeconds: 30
  periodSeconds: 10

readinessProbe:
  httpGet:
    path: /readyz
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 5
```

### Metrics (Coming Soon)

Prometheus metrics endpoint will be available in a future release at `/metrics`.

## 🤝 Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Run `make fmt` before committing
- Ensure `make lint` passes
- Add tests for new features
- Update documentation as needed

## 📝 Roadmap

- [x] Leader election
- [x] Health and readiness endpoints
- [x] Structured logging
- [x] Retry logic with exponential backoff
- [x] Multi-architecture Docker images
- [x] Security hardening (signing, SBOM, CVE scanning)
- [ ] Prometheus metrics
- [ ] Helm chart
- [ ] Automated testing suite
- [ ] IPv6 support
- [ ] Multiple floating IP pools

## 📄 License

This project is a maintained fork of the original
[`cbeneke/hcloud-fip-controller`](https://github.com/cbeneke/hcloud-fip-controller),
which is licensed under the Apache License, Version 2.0.

All modifications in this fork are also provided under the **Apache License, Version 2.0**.

See [LICENSE](LICENSE) for details.

## 🙏 Acknowledgments

- Original project by [Christian Beneke](https://github.com/cbeneke)
- Hetzner Cloud for their excellent API and cloud platform
- The Kubernetes community for the amazing ecosystem

## 📬 Support

- **Issues**: [GitHub Issues](https://github.com/mpowr-it/hetzner-fip-controller/issues)
- **Discussions**: [GitHub Discussions](https://github.com/mpowr-it/hetzner-fip-controller/discussions)
- **Security**: See [SECURITY.md](SECURITY.md) for reporting vulnerabilities

---

**Made with ❤️ by [mpowr](https://github.com/mpowr-it)**