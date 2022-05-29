package github

import "context"

func NewMockGraphQL(tags []string, err error) GraphQL {
	return &mockV4{
		tags: tags,
		err:  err,
	}
}

type mockV4 struct {
	tags []string
	err  error
}

func (m *mockV4) ListTags(ctx context.Context, owner string, repo string) ([]string, error) {
	return m.tags, m.err
}
