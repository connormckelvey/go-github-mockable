.PHONY: clean
clean:
	find . -name "*.gen.go" -type f -delete

.PHONY: client
client:
	go generate .

.PHONY: mocks
mocks:
	go generate ./mocks

.PHONY: gen
gen: clean client mocks

.PHONY: test
test: gen
	go test -v ./...