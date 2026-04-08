# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

The Dedicated Admin Operator manages RBAC permissions for OpenShift Dedicated (OSD) customers. It watches for new namespaces and automatically assigns RoleBindings to the `dedicated-admins` group, while using a regex-based blacklist to protect infrastructure namespaces (kube-*, openshift-*, default, logging, etc.).

## Build Commands

```bash
make gobuild                # Build binary (runs gocheck + gotest first)
make gotest                 # Run all Go tests
make gocheck                # Lint (gofmt -s + go vet)
make build                  # Docker build (requires clean git checkout or ALLOW_DIRTY_CHECKOUT=true)
make generate-syncset       # Generate SelectorSyncSet YAML from manifests/
make clean                  # Remove build artifacts
```

Run a single test package:
```bash
go test -v ./pkg/controller/namespace/
go test -v ./pkg/dedicatedadmin/
```

Build with dirty checkout for local testing:
```bash
ALLOW_DIRTY_CHECKOUT=true make build
```

## Architecture

**Dependency management**: Uses `dep` (Gopkg.toml/Gopkg.lock), not Go modules.

**Operator SDK**: v0.5.0 with controller-runtime v0.1.10. The entrypoint is `cmd/manager/main.go`.

**Two active controllers** (registered via `pkg/controller/add_*.go`):

1. **Namespace Controller** (`pkg/controller/namespace/`): Watches all Namespaces. On create/update, checks the blacklist regex. If not blacklisted, creates two RoleBindings (`dedicated-admins-project-0` and `dedicated-admins-project-1`) binding the `dedicated-admins` group to `dedicated-admins-project` and `admin` ClusterRoles.

2. **RoleBinding Controller** (`pkg/controller/rolebinding/`): Self-healing — if a managed RoleBinding is deleted, it recreates it (unless the namespace is blacklisted).

**Key packages**:
- `config/config.go`: Constants — operator name, namespace (`openshift-dedicated-admin`), default blacklist regex
- `pkg/dedicatedadmin/`: Blacklist logic (`IsBlackListedNamespace`) and operator config loading from ConfigMap
- `pkg/dedicatedadmin/project/resources.go`: RoleBinding definitions (the two bindings created per namespace)
- `pkg/metrics/`: Prometheus gauge for blacklisted namespace count

**Blacklist**: Comma-separated regexes in ConfigMap key `project_blacklist`. Default from `config/config.go`. Evaluated by `dedicatedadmin.IsBlackListedNamespace()`.

**Resource management**: Cluster resources (ClusterRoles, ServiceMonitor, etc.) in `manifests/` are managed by Hive SelectorSyncSet, not by the operator code. The `generate-syncset` make target produces the SyncSet template from manifests.

## Testing

Tests use the standard Go `testing` package with `controller-runtime`'s fake client (`client.NewFakeClient()`). No external test frameworks or envtest setup required.

Test files:
- `pkg/dedicatedadmin/dedicatedadmin_test.go` — blacklist regex matching
- `pkg/controller/namespace/namespace_controller_test.go` — namespace reconciliation
- `pkg/controller/rolebinding/rolebinding_controller_test.go` — RoleBinding self-healing

## Build Outputs

- Binary: `build/_output/bin/dedicated-admin-operator`
- Image: `${IMAGE_REGISTRY}/${IMAGE_REPOSITORY}/dedicated-admin-operator:v${VERSION}` (defaults to `quay.io/$USER/...`)
- SyncSet template: `build/templates/olm-artifacts-template.yaml.tmpl`
