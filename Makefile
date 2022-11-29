TAG = 0.1
TIMESTAMP :=$(shell /bin/date +"%Y%m%d.%H%M%S")
TAG := ghcr.io/basilboli/goskel
COMMIT_HASH := $(shell git log -1 --pretty=format:%h || echo 'N/A')

default:
	go build

.PHONY: all
all: build

.PHONY: test
test:
	go test -v ./...

integration-test:
	docker-compose -f docker-compose.test.yml down -v
	docker-compose -f docker-compose.test.yml build
	docker-compose -f docker-compose.test.yml up --abort-on-container-exit --force-recreate
	docker-compose -f docker-compose.test.yml down -v

clean:
	docker-compose -f docker-compose.test.yml down -v

build:
	docker build --build-arg BUILD_TIME=$(TIMESTAMP) --build-arg COMMIT_HASH=$(COMMIT_HASH) -t $(TAG):$(COMMIT_HASH) .

push:
	docker push $(TAG)

# Note: .env is not included in the repo, but should be created by the user
run-locally:
	. ./.env && go run main.go
