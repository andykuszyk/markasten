build:
	docker build -t andykuszyk/markasten:local .

test:
	go test ./... -v
