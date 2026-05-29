.PHONY: build validate clean

BINARY = infracost-parser-plugin-example

build:
	go build -o $(BINARY) .

validate: build
	infracost plugin validate ./$(BINARY)

clean:
	rm -f $(BINARY)
