# sbomz

SBOM generator for multi-platform container image indexes, inspired by
https://github.com/chainguard-dev/apko.

## Quickstart

### `go install`

```bash
go install github.com/siggy/sbomz@latest
sbomz generate ghcr.io/siggy/sbomz:latest
```

### Pre-built binaries

Download from the [releases page](https://github.com/siggy/sbomz/releases), or
use this script in the repo:

```bash
git clone https://github.com/siggy/sbomz.git && cd sbomz
bin/sbomz generate ghcr.io/siggy/sbomz:latest
```

### Docker

```bash
docker run --rm ghcr.io/siggy/sbomz:latest generate ghcr.io/siggy/sbomz:latest

# for private images
docker run --rm \
  -v $HOME/.docker:/root/.docker:ro -e DOCKER_CONFIG=/root/.docker \
  ghcr.io/siggy/sbomz:latest generate ghcr.io/siggy/sbomz:latest
```

## Usage

```bash
$ sbomz
SBOM generator for multi-platform container image indexes.

Usage:
  sbomz [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  generate    Generate an SBOM from a multi-platform image index
  get         Get an SBOM for a multi-platform image index
  help        Help about any command
  version     Print the version number

Flags:
  -h, --help   help for sbomz

Use "sbomz [command] --help" for more information about a command.
```

### Generate an SBOM, attach it to an image, and verify

```bash
image=[example.com/image:latest]

digest=$(bin/crane digest $image)
image_uri_digest=$image@$digest

sbomz generate $image > sbom.spdx.json

bin/cosign attest \
  --predicate sbom.spdx.json \
  --type spdxjson \
  --yes \
  $image_uri_digest

bin/cosign verify-attestation \
  --type spdxjson \
  --certificate-identity-regexp=.* \
  --certificate-oidc-issuer-regexp=.* \
  $image_uri_digest
```

## Dev

```bash
go run main.go generate ghcr.io/siggy/sbomz:latest
```

## Build

```bash
TAG=$(bin/root-tag)
GO_LDFLAGS="-X sigs.k8s.io/release-utils/version.gitVersion=sbomz-$TAG"

mkdir -p ./target
go build -o ./target/sbomz \
  -ldflags "$GO_LDFLAGS" \
  ./main.go
```

## Docker Build

```bash
TAG=$(bin/root-tag)

docker buildx build --push \
  --platform=linux/amd64,linux/arm64,linux/arm/v7 \
  --build-arg SBOMZ_VERSION=$TAG \
  -t ghcr.io/siggy/sbomz:$TAG \
  -f Dockerfile .
```

## Release

```bash
TAG=v0.0.x
git tag $TAG
git push origin $TAG
```

## Test

```bash
bin/test
```

## Verifying all images and SBOMs

```bash
image=ghcr.io/siggy/sbomz:latest
image_uri_digest=$(bin/crane digest $image --full-ref)

bin/cosign verify \
  --certificate-identity-regexp=.* \
  --certificate-oidc-issuer-regexp=.* \
  $image_uri_digest

bin/cosign verify-attestation \
  --type spdxjson \
  --certificate-identity-regexp=.* \
  --certificate-oidc-issuer-regexp=.* \
  $image_uri_digest

bin/cosign download attestation \
  $image_uri_digest \
  --predicate-type https://spdx.dev/Document |
  bin/jq -r .payload |
  base64 -d |
  bin/jq -r '.predicate.packages[1:][] .versionInfo' |
  while read -r sha; do
    image_uri_digest="$image@$sha"
    bin/cosign verify \
      --certificate-identity-regexp=.* \
      --certificate-oidc-issuer-regexp=.* \
      $image_uri_digest
    bin/cosign verify-attestation \
      --type spdxjson \
      --certificate-identity-regexp=.* \
      --certificate-oidc-issuer-regexp=.* \
      $image_uri_digest
done
```

## TODO

- dev container
- windows support
