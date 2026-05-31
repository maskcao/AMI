.PHONY: run test build tidy

run:
	AMI_ADDR=0.0.0.0:8080 AMI_DATA_DIR=./data AMI_SEED=1 go run ./cmd/ami -seed

test:
	go test ./...

build:
	go build -o bin/ami ./cmd/ami

tidy:
	go mod tidy
