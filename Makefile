build:
	go build -o ./bin/k0porker

run: build
	./bin/k0porker

test:
	go test ./... -v