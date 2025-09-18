# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

The Dedicated Admin Operator is a Kubernetes operator for OpenShift Dedicated that manages permissions via RoleBindings for all client-owned namespaces. It automatically assigns proper permissions to the `dedicated-admins` group when new namespaces are created and maintains these permissions against unauthorized removal.

## Core Components

### Controllers
- **Namespace Controller** (`pkg/controller/namespace/`): Watches for new namespaces and creates RoleBindings for `admin` and `dedicated-admins-project` ClusterRoles to the `dedicated-admins` group
- **RoleBinding Controller** (`pkg/controller/rolebinding/`): Watches for deletion of operator-owned RoleBindings and recreates them
- **Operator Controller**: Manages resources that OLM cannot handle (Service, ServiceMonitor, ClusterRoleBinding)

### Key Packages
- `pkg/dedicatedadmin/`: Core logic for dedicated admin management and blacklist handling
- `pkg/metrics/`: Prometheus metrics exposure
- `config/`: Operator configuration constants
- `cmd/manager/`: Main entry point using operator-sdk framework

## Development Commands

### Building and Testing
```bash
# Build the operator binary
make gobuild

# Run tests
make gotest

# Run Go checks (format, vet)
make gocheck

# Full test including env validation
make test

# Clean build artifacts
make clean
```

### Docker Operations
```bash
# Build Docker image
make build

# Push to registry
make push

# For local testing with dirty checkout
ALLOW_DIRTY_CHECKOUT=true make build
```

### Environment Variables
Key Makefile variables that can be overridden:
- `OPERATOR_NAME`: defaults to value from config/config.go
- `OPERATOR_NAMESPACE`: defaults to value from config/config.go
- `IMAGE_REGISTRY`: defaults to quay.io
- `IMAGE_REPOSITORY`: defaults to $USER
- `ALLOW_DIRTY_CHECKOUT`: set to true for local testing

### SyncSet Generation
```bash
# Generate syncset templates (requires oyaml: pip install oyaml)
make generate-syncset
```

## Architecture Notes

### Operator SDK Framework
This operator uses the operator-sdk framework with controller-runtime. The main entry point sets up:
1. Manager with watch namespace from environment
2. Scheme registration for core and monitoring APIs
3. Controller registration via `controller.AddToManager()`
4. Metrics startup
5. Signal handling for graceful shutdown

### Blacklist Mechanism
The operator uses a blacklist to avoid giving admin permissions to infrastructure/cluster-admin namespaces. This blacklist is loaded from configuration and tracked in metrics.

### Resource Management Philosophy
Per README: Resources are managed externally by Hive via SelectorSyncSet rather than bundled with OLM, as OLM cannot properly update or delete bundled resources. This simplifies deployment and ensures proper reconciliation.

### Manual Testing Workflow
For testing new operator versions:
1. Remove OLM subscription if present
2. Build and push image with custom repository
3. Deploy manifests with updated image reference
4. Validate functionality

## Key Files
- `Makefile`: Main build system entry point
- `project.mk`: Project-specific configuration
- `standard.mk`: Standard build targets and variables
- `manifests/`: Kubernetes resource definitions
- `cmd/manager/main.go`: Operator entry point