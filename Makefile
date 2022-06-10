
run:
	go run .

test:
	go test ./...

coverage:
	go test -json -coverprofile=cover.out ./... > result.json
	go tool cover -func cover.out
	go tool cover -html=cover.out

fmt:
	go fmt ./...

tidy:
	go mod tidy