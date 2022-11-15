package gogithubmockable_test

import (
	"context"
	"testing"

	gogithubmockable "github.com/connormckelvey/go-github-mockable"
	"github.com/google/go-github/v48/github"
	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	g := github.NewClient(nil)
	var c gogithubmockable.ClientAPI
	c = gogithubmockable.NewClient(g)
	c.Repositories().Get(context.TODO(), "owner", "repo")
	_, ok := c.(*gogithubmockable.Client)
	assert.True(t, ok)
}
