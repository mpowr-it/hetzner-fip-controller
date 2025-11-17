# hcloud-fip-controller (maintained fork)

This repository contains a maintained fork of the original
[`cbeneke/hcloud-fip-controller`](https://github.com/mpowr/hetzner-fip-controller).

The controller manages Hetzner Cloud Floating IPs (FIPs) and assigns them to
Kubernetes nodes based on labels and pod IPs. It is typically deployed on
control plane nodes in Hetzner Cloud based Kubernetes clusters.

## Why this fork?

The original project is no longer actively maintained and shows some issues in
modern clusters and with transient Hetzner Cloud API outages:

- the controller terminates on temporary `503 service unavailable` responses
  from the Hetzner API, which causes unnecessary pod restarts
- there are no health or readiness endpoints
- logging is limited and does not provide enough detail for debugging

This fork aims to:

- make the controller more robust against transient Hetzner API issues
- improve observability (logging, health/readiness probes, later metrics)
- keep changes minimal and focused, so it can be dropped in as a replacement

## Features

- Leader Election to ensure only one active controller instance
- Periodic reconciliation of Floating IP assignments based on pod and node IPs
- Configurable via ConfigMap and Kubernetes Secrets
- Runs inside the cluster using in-cluster configuration

Planned improvements in this fork:

- no `log.Fatal` on temporary Hetzner API errors
- simple retry logic with backoff for failed reconciliations
- HTTP `/healthz` and `/readyz` endpoints for Kubernetes probes
- optional Prometheus metrics (in a later step)

## Getting started

### Building

```bash
go build ./cmd/hcloud-fip-controller

## License

This project is a maintained fork of the original
[cbeneke/hcloud-fip-controller](https://github.com/mpowr/hetzner-fip-controller),
which is licensed under the Apache License, Version 2.0.

All modifications in this fork are also provided under the Apache 2.0 License.

See LICENSE for details.
