package gogithubmockable_test

import (
	"testing"

	gogithubmockable "github.com/connormckelvey/go-github-mockable"
	"github.com/google/go-github/v48/github"
	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	g := github.NewClient(nil)
	var c gogithubmockable.ClientAPI
	c = gogithubmockable.NewClient(g)
	_, ok := c.(*gogithubmockable.Client)
	assert.True(t, ok)
}
