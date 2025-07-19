package main

import (
	"io"
	"os"

	"github.com/awslabs/amazon-ecr-credential-helper/ecr-login"
	"github.com/chrismellard/docker-credential-acr-env/pkg/credhelper"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/authn/github"
	"github.com/google/go-containerregistry/pkg/v1/google"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/siggy/sbomz/cmd"
)

func main() {
	amazonKeychain := authn.NewKeychainFromHelper(ecr.NewECRHelper(ecr.WithLogger(io.Discard)))
	azureKeychain := authn.NewKeychainFromHelper(credhelper.NewACRCredentialsHelper())

	keychain := authn.NewMultiKeychain(
		authn.DefaultKeychain,
		google.Keychain,
		github.Keychain,
		amazonKeychain,
		azureKeychain,
	)

	remoteOpts := remote.WithAuthFromKeychain(keychain)

	err := cmd.NewRootCmd(remoteOpts).Execute()
	if err != nil {
		os.Exit(1)
	}
}
