build:
	go build -o ./bin/k0porker

run: build
	./bin/k0porker

test:
	go test ./... -v

proto:
	protoc --go_out=. --go_opt=paths=source_relative \
	--go-grpc_out=. --go-grpc_opt=paths=source_relative \
	proto/service.proto

.PHONY: proto