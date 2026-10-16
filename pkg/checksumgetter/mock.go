package checksumgetter

import (
	"context"
	"log/slog"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/runtime"
)

type MockChecksumFileVerifier struct {
	Err error
	// Kinds is what VerifyAsset and VerifyAssetDigest report as verified.
	Kinds []string
	// DigestOK is whether VerifyAssetDigest could verify the digest.
	DigestOK bool
	// Digests records the digests VerifyAssetDigest was given.
	Digests []string
}

func (v *MockChecksumFileVerifier) VerifyChecksumFileContent(_ context.Context, _ *slog.Logger, _ *config.Package, _ string, _ []byte) error {
	return v.Err
}

func (v *MockChecksumFileVerifier) VerifyAsset(_ context.Context, _ *slog.Logger, _ *config.Package, _, _ string, _ *runtime.Runtime) ([]string, error) {
	return v.Kinds, v.Err
}

func (v *MockChecksumFileVerifier) VerifyAssetDigest(_ context.Context, _ *slog.Logger, _ *config.Package, _, digest string, _ *runtime.Runtime) ([]string, bool, error) {
	v.Digests = append(v.Digests, digest)
	if v.Err != nil || !v.DigestOK {
		return nil, false, v.Err
	}
	return v.Kinds, true, nil
}
