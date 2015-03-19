.PHONY: all test fmt benchmark git-add-hook

all: test fmt vet

vet:
	go vet ./...

build:
	go build ./bin/logmunch

clean:
	rm -f logmunch
	go clean ./...

test:
	go test ./...

fmt:
	go fmt ./...

benchmark:
	go test ./... -bench=".*"

git-pre-commit-hook:
	curl -s 'http://tip.golang.org/misc/git/pre-commit?m=text' > .git/hooks/pre-commit
	chmod +x .git/hooks/pre-commit
