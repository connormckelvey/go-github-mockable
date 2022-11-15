# Go Github Mockable

[![Go Reference](https://pkg.go.dev/badge/github.com/connormckelvey/go-github-mockable.svg)](https://pkg.go.dev/github.com/connormckelvey/go-github-mockable) [![Go Report Card](https://goreportcard.com/badge/github.com/connormckelvey/go-github-mockable)](https://goreportcard.com/report/github.com/connormckelvey/go-github-mockable)

Go Github Mockable provides an interface-based wrapper for the [Go Github](https://github.com/google/go-github) client, making it possible to generate mocks (also included) for the client using [GoMock](https://github.com/golang/mock). 

## Installation

```bash
$ go get github.com/connormckelvey/go-github-mockable
```

## Features

### [ClientAPI interface type](https://pkg.go.dev/github.com/connormckelvey/go-github-mockable#ClientAPI) 

Interface type to be used in place of `*github.Client`. Allows dependency injection of a mocked `ClientAPI`. 

For example

```go

type MyService struct {
    github *github.Client
}

func NewMyService(client *github.Client) *Service {
    return &MyService{
        github: client,
    }
}

//becomes:

type MyService struct {
    github gogithubmockable.ClientAPI
}

func NewMyService(client gogithubmockable.ClientAPI) *Service {
    return &MyService{
        github: client,
    }
}
```

### [ClientAPI implementation](https://pkg.go.dev/github.com/connormckelvey/go-github-mockable#Client)

`gogithubmockable.Client` provides a default implementation of `gogithubmockable.ClientAPI` by dispatching calls to a provided `*github.Client` 

For example:

```go

func main() {
    gh := github.NewClient(nil)
    client := gogithubmockable.NewClient(gh)

    service := NewMyService(client)
    ...
}


func TestMyService(t *testing.T) {
    ctrl := gomock.NewController(t)

	mockClient := mocks.NewMockClientAPI(ctrl)
    service := NewMYService(mockClient)
}
```


### [Services Getter Methods](https://pkg.go.dev/github.com/connormckelvey/go-github-mockable#Client.Actions)

The `ClientAPI` interface include getter methods for every service normally available on the `*github.Client` struct. The return types for the getter methods are themselves interfaces allowing services to be independently mocked. 

For example

```go
func (s *MyService) Foo() {
    repo, _, err := s.github.Repositories.Get(context.Background(), "owner", "repo")
    ...
}

// becomes

func (s *MyService) Foo() {
    repo, _, err := s.github.Repositories().Get(context.Background(), "owner", "repo")
    ...
}
```

### [Service interfaces](https://pkg.go.dev/github.com/connormckelvey/go-github-mockable#ActionsService)

The service getter methods included on the `ClientAPI` interface return interface types defined with all public methods included on the original concrete service types. 



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

