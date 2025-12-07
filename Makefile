generate:
	go build -o bin/protoc-gen-proterror ./cmd/protoc-gen-proterror/main.go && buf generate

linter:
	golangci-lint --config .golangci.yaml run

fmt:
	golangci-lint --config .golangci.yaml fmt
