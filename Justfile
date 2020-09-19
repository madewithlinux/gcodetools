fmt:
    go fmt -x ./...

test:
    go test -v ./...

cover:
    go test -coverprofile=coverage.out
    go tool cover -html=coverage.out

ci:
    #!/bin/bash
    diff -u <(echo -n) <(gofmt -d .)
    go vet ./...
    go test -v -race ./...
