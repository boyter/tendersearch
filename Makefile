# Some people have gotestsum installed and like it so use it if it exists
HAS_GOTESTSUM := $(shell which gotestsum)
ifdef HAS_GOTESTSUM
    TEST_CMD = gotestsum --format testname --packages="./..." -- -count=1 -tags=integration -v -p 1
else
    TEST_CMD = go test ./... --count=1 -tags=integration
endif

lint: gitleaks
	@golangci-lint run --disable-all --enable gci --fix
	@golangci-lint run

gitleaks:
	gitleaks detect -v -c gitleaks.toml --log-opts "-p -n 100"

test:
	@$(TEST_CMD)

test-run:
	@$(TEST_CMD) -run=$(RUN)

test-coverage:
	go test ./... -coverprofile coverage.out && go tool cover -html=coverage.out -o coverage.html

test-smoke:
	go run tasks.go -smoke-test

check-unique-code:
	 go run tasks.go -check-unique

confirm-unique-code:
	 go run tasks.go -confirm-unique

gen: sqlcgen
	go generate -v ./...

sqlcgen:
	sqlc generate

mod:
	@go mod tidy
	@go mod vendor

clean:
	go clean -modcache

web-run:
	@cd cmd/tendersearch && go build . && ./tendersearch

smoke-run:
	@cd cmd/api-web && go build . && HTTP_SERVER_PORT=4001 ./api-web

# Build all binaries assuming more than one in the future
build:
	cd cmd/tendersearch && go build

all: mod gen build lint test test-smoke check-unique-code confirm-unique-code
