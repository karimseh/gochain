
`Makefile`
```makefile
BIN_NAME = gochain

build:
	go build -o bin/$(BIN_NAME) cmd/gochain/main.go

test:
	go test -v ./...

clean:
	rm -rf bin/