BIN=dt-geo-converter
VERSION?=testing

.PHONY: build vet test clean

build:
	go build -ldflags "-X dt-geo-converter/cmd.Version=$(VERSION)" -o $(BIN) .

vet:
	go vet ./...

clean:
	rm -f $(BIN)
