package mocks

//go:generate go run -mod=mod -workfile=off github.com/golang/mock/mockgen@v1.6.0 -source=../client.gen.go -package=mocks -destination mocks.gen.go
