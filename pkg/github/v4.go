package github

import (
	"context"
	"fmt"

	"github.com/shurcooL/githubv4"
)

type GraphQL interface {
	ListTags(ctx context.Context, owner string, repo string) ([]string, error)
}

func NewGraphQL(v4Client *githubv4.Client) GraphQL {
	return &graphQL{
		client: v4Client,
	}
}

type graphQL struct {
	client *githubv4.Client
}

func (cl *graphQL) ListTags(ctx context.Context, owner string, repo string) ([]string, error) {
	// Pagination isn't supported
	var q struct {
		Repository struct {
			Refs struct {
				Nodes []struct {
					Name string
				}
			} `graphql:"refs(refPrefix: \"refs/tags/\", first:30, orderBy:{direction:DESC, field:TAG_COMMIT_DATE})"`
		} `graphql:"repository(owner: $owner, name: $name)"`
	}
	variables := map[string]interface{}{
		"owner": githubv4.String(owner),
		"name":  githubv4.String(repo),
	}

	if err := cl.client.Query(ctx, &q, variables); err != nil {
		return nil, fmt.Errorf("list tags by GitHub GraphQL API: %w", err)
	}
	tags := make([]string, len(q.Repository.Refs.Nodes))
	for i, node := range q.Repository.Refs.Nodes {
		tags[i] = node.Name
	}
	return tags, nil
}
