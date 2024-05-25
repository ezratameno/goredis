tidy:
	@go mod tidy

run: build
	@./bin/goredis

build:
	@ go build -o bin/goredis .

telnet:
	@telnet localhost 5001

	# ctl + ]
	# then 'q'