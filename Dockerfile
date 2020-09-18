# Build the mattermost operator
ARG BUILD_IMAGE=golang:1.14
ARG BASE_IMAGE=gcr.io/distroless/static:nonroot

FROM ${BUILD_IMAGE} as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/
COPY pkg/ pkg/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o manager main.go

FROM ${BASE_IMAGE}

LABEL name="Mattermost Operator" \
  maintainer="dev-ops@mattermost.com" \
  vendor="Mattermost" \
  distribution-scope="public" \
  architecture="x86_64" \
  url="https://mattermost.dev" \
  io.k8s.description="Mattermost Operator creates, configures and helps manage Mattermost installations on Kubernetes" \
  io.k8s.display-name="Mattermost Operator" \
  io.openshift.tags="mattermost,collaboration,operator" \
  summary="Quick and easy Mattermost setup" \
  description="Mattermost operator deploys and configures Mattermost installations, and assists with maintenance/upgrade operations."


WORKDIR /
COPY --from=builder /workspace/manager .
USER nonroot:nonroot

ENTRYPOINT ["/manager"]
