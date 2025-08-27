.DEFAULT_GOAL := build

BIN_FILE=payment-dates-advisor

build:
	@go build -o ${BIN_FILE} ./cmd/

clean:
	rm -f "${BIN_FILE}"
	rm -f "cp.out"
	rm -f nohup.out
	rm -f "${BIN_FILE}"

test:
	go test -v ./...

check:
	go test -v ./...

cover:
	go test ./... -coverprofile=cp.out
	go tool cover -html=cp.out

run:
	./"${BIN_FILE}"

lint:
	golangci-lint run
