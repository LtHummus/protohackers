.PHONY: build clean linux

build:
	go build -o protohackers main.go

linux:
	GOOS=linux GOARCH=amd64 go build -o protohackers main.go

clean:
	rm protohackers