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


    // Instead of client.Repositories, use client.GetRepositories()
    c.GetRepositories().Get(context.TODO(), "owner", "repo")
    ...
}
```

## TODO

- Copy remaining util methods on *github.Client (not services or opts)
