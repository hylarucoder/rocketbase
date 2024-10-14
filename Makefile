lint:
	golangci-lint run -c ./golangci.yml ./...

test:
	go test ./... -v --cover

test-fast:
	go test -failfast -v ./... 

jstypes:
	go run ./plugins/jsvm/internal/types/types.go

test-report:
	go test ./... -v --cover -coverprofile=coverage.out
	go tool cover -html=coverage.out
