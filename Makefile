version?=0.1.0
type?=alpha
build?=0
commit=$(shell git rev-parse --short HEAD)
package=github.com/open-osquery/trailsc
LDFLAGS=-X $(package)/internal/config.Version=$(version)
LDFLAGS+= -X $(package)/internal/config.Release=$(type)
LDFLAGS+= -X $(package)/internal/config.Commit=$(commit)
LDFLAGS+= -X $(package)/internal/ingestd.Build=$(build)

install:
	go install -ldflags "-s -w $(LDFLAGS)"

build:
	go build -ldflags "-s -w $(LDFLAGS)" -o trailsc

test:
	go test -v ./...

pre-commit: FORCE
	# Setup an alias "gitos" that has your opensource git profile configured.
	# atleast user.name and user.email
	@gitos
	@go test ./...
	@go fmt ./...
	@goimports -w
	@golint ./...
	@go vet ./...
	@go mod tidy

deps:
	@echo "Installing tools goimports, golint, gocyclo"
	@go get -u golang.org/x/lint/golint
	@go get -u github.com/fzipp/gocyclo
	@go get golang.org/x/tools/cmd/goimports
	@echo "Setting up pre-commit hook"
	@ln -snf ../../.pre-commit .git/hooks/pre-commit
	@chmod +x .pre-commit

FORCE: ;
