package cmd

import (
	"fmt"
	"time"

	"chainguard.dev/apko/pkg/build/types"
	"chainguard.dev/apko/pkg/sbom"
	"chainguard.dev/apko/pkg/sbom/generator/spdx"
	"chainguard.dev/apko/pkg/sbom/options"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	ggcrtypes "github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/spf13/cobra"
)

func newGenerateCmd(opts remote.Option) *cobra.Command {
	var timestamp int64

	cmd := &cobra.Command{
		Use:   "generate IMAGE [flags]",
		Short: "Generate an SBOM from a multi-platform image index",
		Args:  cobra.ExactArgs(1),
		Long:  `Generate an SPDX SBOM from a multi-platform container image index.`,
		Example: `
# Generate an SBOM
sbomz generate ghcr.io/siggy/sbomz:latest

# Generate an SBOM with a specific timestamp
sbomz generate --timestamp $(date "+%s") ghcr.io/siggy/sbomz:latest`,
		RunE: func(cmd *cobra.Command, args []string) error {
			image := args[0]

			t := time.Now()
			if cmd.Flags().Changed("timestamp") {
				t = time.Unix(timestamp, 0)
			}

			return generate(image, t, opts)
		},
	}

	cmd.Flags().Int64VarP(&timestamp, "timestamp", "t", 0, "unix timestamp to put on the SBOM (default: current time)")

	return cmd
}

// generate fetches the image index, extracts the digest and images, and prints
// an SPDX SBOM to stdout.
func generate(image string, t time.Time, remoteOpts remote.Option) error {
	digest, images, err := fetchInfo(image, remoteOpts)
	if err != nil {
		return fmt.Errorf("getting image info: %w", err)
	}

	opts := mkOptions(digest, images, t)

	gen := &spdx.SPDX{}
	return gen.GenerateIndex(&opts, "/dev/stdout")
}

func fetchInfo(imageRef string, remoteOpts remote.Option) (v1.Hash, []options.ArchImageInfo, error) {
	ref, err := name.ParseReference(imageRef)
	if err != nil {
		return v1.Hash{}, nil, fmt.Errorf("parse image reference %q: %w", imageRef, err)
	}

	idx, err := remote.Index(ref, remoteOpts)
	if err != nil {
		return v1.Hash{}, nil, fmt.Errorf("fetch index for %q: %w", imageRef, err)
	}

	digest, err := idx.Digest()
	if err != nil {
		return v1.Hash{}, nil, fmt.Errorf("get index digest for %q: %w", imageRef, err)
	}

	indexManifest, err := idx.IndexManifest()
	if err != nil {
		return v1.Hash{}, nil, fmt.Errorf("get index manifest for %q: %w", imageRef, err)
	}

	images := []options.ArchImageInfo{}
	for _, desc := range indexManifest.Manifests {
		if desc.Platform.OS == "unknown" && desc.Platform.Architecture == "unknown" {
			continue
		}

		images = append(images, options.ArchImageInfo{
			Digest: desc.Digest,
			Arch:   types.Architecture(desc.Platform.Architecture),
		})
	}

	return digest, images, nil
}

func mkOptions(digest v1.Hash, images []options.ArchImageInfo, t time.Time) options.Options {
	opts := sbom.DefaultOptions
	opts.ImageInfo.Images = images
	opts.ImageInfo.IndexDigest = digest
	opts.ImageInfo.IndexMediaType = ggcrtypes.OCIImageIndex
	opts.ImageInfo.SourceDateEpoch = t
	opts.OS.Name = "sbomz"

	return opts
}
