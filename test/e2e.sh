#!/usr/bin/env bash

set -Eeuxo pipefail

readonly REPO_ROOT="${REPO_ROOT:-$(git rev-parse --show-toplevel)}"

run_ct_container() {
    echo 'Running ct container...'
    docker run --rm --interactive --detach --network host --name test-cont \
        --volume "$(pwd):/go/src/github.com/mattermost/mattermost-operator" \
        --workdir "/go/src/github.com/mattermost/mattermost-operator" \
        "golang:1.12.2" \
        cat
    echo
}

docker_exec() {
    docker exec --interactive test-cont "$@"
}


run_kind() {
    echo "Download kind binary..."
    curl -sSLo kind https://github.com/kubernetes-sigs/kind/releases/download/"${KIND_VERSION}"/kind-linux-amd64
    chmod +x kind
    sudo mv kind /usr/local/bin/kind

    kind --version

    echo "Download kubectl..."
    curl -sSLo kubectl https://storage.googleapis.com/kubernetes-release/release/"${K8S_VERSION}"/bin/linux/amd64/kubectl
    chmod +x kubectl
    sudo cp kubectl /usr/local/bin/
    docker cp kubectl test-cont:/usr/local/bin/
    echo

    echo "Create Kubernetes cluster with kind..."
    kind create cluster --config test/kind-config.yaml --wait 5m

    echo "Export kubeconfig..."
    # shellcheck disable=SC2155
    export KUBECONFIG="$(kind get kubeconfig-path)"
    cp "$(kind get kubeconfig-path)" ~/.kube/config
    echo

    echo 'Copying kubeconfig to container...'
    local kubeconfig
    kubeconfig="$(kind get kubeconfig-path)"
    docker_exec mkdir /root/.kube
    docker cp "$kubeconfig" test-cont:/root/.kube/config
    docker_exec kubectl cluster-info
    echo

    echo -n 'Waiting for cluster to be ready...'
    until ! grep --quiet 'NotReady' <(kubectl get nodes --no-headers); do
        printf '.'
        sleep 5
    done

    kubectl get all --all-namespaces
}

install_operator-sdk() {
    echo "Install operator-sdk"
    MACHINE="$(uname -m)"
    curl -Lo build/operator-sdk https://github.com/operator-framework/operator-sdk/releases/download/"${SDK_VERSION}"/operator-sdk-"${SDK_VERSION}"-"${MACHINE}"-linux-gnu
    chmod +x build/operator-sdk
    docker cp build/operator-sdk test-cont:/usr/local/bin/
    echo
}

cleanup() {
    echo 'Removing test container...'
    docker kill test-cont > /dev/null 2>&1
    echo 'Removing Kind Cluster...'
    kind delete cluster

    echo 'Done!'
}

main() {
    run_ct_container
    trap cleanup EXIT

    run_kind

    install_operator-sdk

    echo "Ready for testing"

    # Build the operator container image.
    # This would build a container with tag mattermost/mattermost-operator:test,
    # which is used in the e2e test setup below.
    make build-image

    # Move the operator container inside Kind container so that the image is
    # available to the docker in docker environment.
    # Copy the image to the cluster to make a bit more fast to start
    docker pull iad.ocir.io/oracle/mysql-operator:0.3.0
    kind load docker-image iad.ocir.io/oracle/mysql-operator:0.3.0
    kind load docker-image mattermost/mattermost-operator:test

    # Create a namespace for testing operator.
    # This is needed because the service account created using
    # deploy/service_account.yaml has a static namespace. Creating operator in
    # other namespace will result in permission errors.
    kubectl create ns mattermost-operator

    # Create the mysql operator
    kubectl create ns mysql-operator
    kubectl apply -n mysql-operator -f test/mysql/crds/mysql_crd.yaml
    kubectl apply -n mysql-operator -f test/mysql/service_account.yaml
    kubectl apply -n mysql-operator -f test/mysql/role.yaml
    kubectl apply -n mysql-operator -f test/mysql/role_binding.yaml
    kubectl apply -n mysql-operator -f test/mysql/operator.yaml

    # NOTE: Append this test command with `|| true` to debug by inspecting the
    # resource details. Also comment `defer ctx.Cleanup()` in the cluster to
    # avoid resouce cleanup.
    echo "Starting Operator Testing..."
    docker_exec operator-sdk test local ./test/e2e --namespace mattermost-operator --kubeconfig /root/.kube/config

    echo "Done Testing!"
}

main "$@"