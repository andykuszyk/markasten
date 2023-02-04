TAG = local
ifneq ($(GITHUB_REF_NAME),)
	TAG = $(GITHUB_REF_NAME)
endif

build:
	docker build -t andykuszyk/markasten:${TAG} .

login:
	docker login -u "$(DOCKER_USERNAME)" -p "$(DOCKER_PASSWORD)"

push:
	docker push andykuszyk/markasten:${TAG}

test:
	go test ./... -v
