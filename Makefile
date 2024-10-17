build:
	go build -o bin/wc . 
run:
	./bin/wc -rps 2 -workers 8 -n 10 -progress 10 -batch-size 4000
help:
	./bin/wc --help
br: build run
bh: build help

lint:
	golangci-lint run
test:
	go test
