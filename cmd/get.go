package cmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/spf13/cobra"
)

func newGetCmd(opts remote.Option) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get IMAGE [flags]",
		Short: "Get an SBOM for a multi-platform image index",
		Args:  cobra.ExactArgs(1),
		Long:  `Get an SPDX SBOM for a multi-platform container image index.`,
		Example: `
# Get an SBOM
sbomz get ghcr.io/siggy/sbomz:latest`,
		RunE: func(cmd *cobra.Command, args []string) error {
			image := args[0]
			return get(image, opts)
		},
	}

	return cmd
}

// get fetches the image index's attestation, and prints to stdout.
func get(image string, remoteOpts remote.Option) error {
	sbom, err := fetchSBOM(image, remoteOpts)
	if err != nil {
		return fmt.Errorf("getting SBOM: %w", err)
	}

	sbomJSON, err := json.MarshalIndent(sbom, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling SBOM to JSON: %w", err)
	}

	fmt.Printf("%s\n", sbomJSON)

	return nil
}

func fetchSBOM(imageRef string, remoteOpts remote.Option) (json.RawMessage, error) {
	ref, err := name.ParseReference(imageRef)
	if err != nil {
		return nil, fmt.Errorf("parse image reference %q: %w", imageRef, err)
	}

	idx, err := remote.Index(ref, remoteOpts)
	if err != nil {
		return nil, fmt.Errorf("fetch index for %q: %w", imageRef, err)
	}

	digest, err := idx.Digest()
	if err != nil {
		return nil, fmt.Errorf("get index digest for %q: %w", imageRef, err)
	}

	attTag := fmt.Sprintf("%s:%s-%s.att", ref.Context(), digest.Algorithm, digest.Hex)

	attName, err := name.ParseReference(attTag)
	if err != nil {
		return nil, fmt.Errorf("parse image reference %q: %w", imageRef, err)
	}

	attImg, err := remote.Image(attName, remoteOpts)
	if err != nil {
		return nil, fmt.Errorf("fetch index for %q: %w", imageRef, err)
	}

	sbom, err := extractSBOMFromImage(attImg)
	if err != nil {
		return nil, fmt.Errorf("extract SBOM from image %q: %w", attImg, err)
	}

	return sbom, nil
}

func extractSBOMFromImage(img v1.Image) (json.RawMessage, error) {
	layers, err := img.Layers()
	if err != nil {
		return nil, err
	}

	for _, layer := range layers {
		r, err := layer.Uncompressed()
		if err != nil {
			continue
		}
		defer func() {
			err := r.Close()
			if err != nil {
				fmt.Printf("error closing layer reader: %v\n", err)
			}
		}()

		data, err := io.ReadAll(r)
		if err != nil {
			continue
		}

		var env struct {
			Payload string `json:"payload"`
		}
		if err := json.Unmarshal(data, &env); err != nil {
			continue
		}

		decoded, err := base64.StdEncoding.DecodeString(env.Payload)
		if err != nil {
			continue
		}

		var stmt struct {
			Predicate     json.RawMessage `json:"predicate"`
			PredicateType string          `json:"predicateType"`
		}
		err = json.Unmarshal(decoded, &stmt)
		if err != nil {
			continue
		}

		if stmt.PredicateType == "https://spdx.dev/Document" {
			return stmt.Predicate, nil
		}
	}

	return nil, fmt.Errorf("no valid SBOM found in image layers: %s", img)
}
