# Precompile key slow-to-build dependencies
FROM --platform=$BUILDPLATFORM golang:1.24.4-alpine AS go-deps

RUN apk update
RUN apk add --no-cache ca-certificates \
    && update-ca-certificates

WORKDIR /build
COPY go.mod go.sum ./
COPY main.go main.go
COPY cmd cmd

RUN go mod vendor

## compile sbomz CLI
FROM --platform=$BUILDPLATFORM go-deps AS golang
WORKDIR /build

ARG SBOMZ_VERSION
ENV GO_LDFLAGS="-s -w -X sigs.k8s.io/release-utils/version.gitVersion=sbomz-${SBOMZ_VERSION}"

ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
  go build -o /out/sbomz -mod=vendor -ldflags "${GO_LDFLAGS}" ./

## package runtime
FROM scratch
LABEL org.opencontainers.image.source=https://github.com/siggy/sbomz
LABEL org.opencontainers.image.description="SBOM generator for multi-platform container image indexes"
COPY --from=golang /out/sbomz /sbomz
COPY --from=golang /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT ["/sbomz"]
