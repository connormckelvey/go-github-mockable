package mocks_test

import (
	"context"
	"fmt"
	"testing"

	gogithubmockable "github.com/connormckelvey/go-github-mockable"
	"github.com/connormckelvey/go-github-mockable/mocks"
	gomock "github.com/golang/mock/gomock"
	github "github.com/google/go-github/v48/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MyService struct {
	Github gogithubmockable.ClientAPI
}

func NewMyService(github gogithubmockable.ClientAPI) *MyService {
	return &MyService{
		Github: github,
	}
}

func (s *MyService) FullName() (string, error) {
	r, _, err := s.Github.Repositories().Get(context.Background(), "connormckelvey", "go-github-mockable")
	return *r.FullName, err
}

func TestMockClientAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	owner := "connormckelvey"
	repo := "go-github-mockable"
	expected := fmt.Sprintf("%s/%s", owner, repo)

	rm := mocks.NewMockRepositoriesService(ctrl)
	rm.EXPECT().Get(gomock.Any(), gomock.Eq(owner), gomock.Eq(repo)).Return(
		&github.Repository{
			FullName: &expected,
		},
		&github.Response{},
		nil,
	)

	cm := mocks.NewMockClientAPI(ctrl)
	cm.EXPECT().Repositories().Return(rm)

	service := NewMyService(cm)
	fullName, err := service.FullName()
	require.NoError(t, err)
	assert.Equal(t, expected, fullName)
}
