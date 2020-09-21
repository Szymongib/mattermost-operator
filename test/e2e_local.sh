#!/usr/bin/env bash

## Run e2e tests on local machine
## Requirement:
## - kind 0.9.0
## - kustomize

set -Eeuxo pipefail

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

CLUSTER_NAME=${CLUSTER_NAME:-kind}

# Create Kind cluster if does not exist
if [[ $(kind get clusters | grep ${CLUSTER_NAME}) ]]; then
    echo "Kind cluster already exist."
else
    echo "Creating Kind cluster"
    kind create cluster --config ${DIR}/kind-config.yaml --name ${CLUSTER_NAME}
fi

# Use Kind cluster kubeconfig
kind export kubeconfig --name ${CLUSTER_NAME}

source ${DIR}/setup_test.sh

go test ./test/e2e --timeout 30m
