# Go Github Mockable

Go Github Mockable provides an interface-based wrapper for the [Go Github](https://github.com/google/go-github) client, making it possible to generate mocks (also included) for the client using [GoMock](https://github.com/golang/mock). 

## Install

```bash
$ go get github.com/connormckelvey/go-github-mockable
```

## Usage

```go
package main

import (
    "github.com/google/go-github/v48/github"
    "github.com/connormckelvey/go-github-mockable"
)

func main() {
    gh := github.NewClient(nil)
    client := gogithubmockable.NewClient(gh)


    // Instead of client.Repositories, use client.Repositories()
    c.Repositories().Get(context.TODO(), "owner", "repo")
    ...
}
```

### Mocking

```go
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
```

