.DEFAULT_GOAL := build
BIN_FILE=bookings
CMD_DIR=./cmd/web
build:
	@go build -o "${BIN_FILE}" $(CMD_DIR)/*.go 
clean:
	go clean ./...
	rm --force "cp.out"
	rm --force nohup.out
test:
	go test -v ./...
check:
	go test -v ./...
cover:
	go test ./... -coverprofile cp.out
	go tool cover -html=cp.out
run:
	./"${BIN_FILE}"
lint:
	golangci-lint run --enable-all
