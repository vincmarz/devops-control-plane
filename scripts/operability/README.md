# Operability scripts

This directory contains read-only operational scripts for the DevOps Control Plane.

## `health-check.sh`

Smoke-test script for Phase 10.3.3.

The script performs a read-only health-check against the DevOps Control Plane runtime on OpenShift and saves an evidence package under `/tmp` by default.

### Safety

The script does not:

- print Secret values;
- decode Secret values;
- apply, patch, edit, delete or restart resources.

### Default execution

```bash
./scripts/operability/health-check.sh
```

### Optional environment variables

```bash
export DCP_NS="devops-control-plane"
export APP_LABEL="app=devops-control-plane"
export PG_LABEL="app=postgresql"
export EVIDENCE_DIR="/tmp/dcp-operability-smoke-test-$(date +%Y%m%d-%H%M%S)"
export SKIP_CAN_I="false"

./scripts/operability/health-check.sh
```

### Exit codes

- `0`: all mandatory checks passed
- `1`: at least one mandatory check failed
- `2`: local prerequisite or OpenShift context problem
