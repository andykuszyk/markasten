TAG = local
ifneq ($(GITHUB_REF_NAME),)
	TAG = $(GITHUB_REF_NAME)
endif

build:
	docker build -t form3tech/markasten:${TAG} .

login:
	docker login -u "$(DOCKER_USERNAME)" -p "$(DOCKER_PASSWORD)"

push:
	docker push form3tech/markasten:${TAG}
	docker tag form3tech/markasten:${TAG} form3tech/markasten:latest
	docker push form3tech/markasten:latest

test:
	go test ./... -v
