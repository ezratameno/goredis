tidy:
	@go mod tidy

run: build
	@./bin/goredis

build:
	@go build -o bin/goredis .

test:
	@ go test -count=1 ./... 
telnet:
	@telnet localhost 5001

	# ctl + ]
	# then 'q'