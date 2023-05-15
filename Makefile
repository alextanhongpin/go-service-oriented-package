install:
	@go install github.com/vektra/mockery/v2@v2.20.0


gen:
	@go generate ./...


test:
	@go test -v -cover -coverprofile=cover.out ./...
	@go tool cover -html=cover.out
