package gogithubmockable_test

import (
	"context"
	"testing"

	gogithubmockable "github.com/connormckelvey/go-github-mockable"
	"github.com/google/go-github/v48/github"
	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	gh := github.NewClient(nil)

	var c gogithubmockable.ClientAPI = gogithubmockable.NewClient(gh)
	_, ok := c.(*gogithubmockable.Client)
	assert.True(t, ok)

	r, _, err := c.Repositories().Get(context.Background(), "connormckelvey", "go-github-mockable")
	assert.NoError(t, err)

	assert.Equal(t, "go-github-mockable", *r.Name)
}
