.PHONY: build validate clean

BINARY = infracost-provider-plugin-example

build:
	go build -o $(BINARY) .

validate: build
	infracost plugin validate ./$(BINARY)

clean:
	rm -f $(BINARY)
