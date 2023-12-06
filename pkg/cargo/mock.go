package cargo

import (
	"context"
)

type MockClient struct {
	Versions     []string
	Err          error
	CratePayload *CratePayload
}

func (m *MockClient) ListVersions(_ context.Context, _ string) ([]string, error) {
	return m.Versions, m.Err
}

func (m *MockClient) GetLatestVersion(_ context.Context, _ string) (string, error) {
	if len(m.Versions) == 0 {
		return "", m.Err
	}
	return m.Versions[0], m.Err
}

func (m *MockClient) GetCrate(_ context.Context, _ string) (*CratePayload, error) {
	return m.CratePayload, m.Err
}
