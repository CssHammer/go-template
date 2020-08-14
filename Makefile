SHELL=/bin/bash
ROOT_DIR := $(shell pwd)
IMAGE_TAG := $(shell git rev-parse --short HEAD)
IMAGE_NAME := company/srv
REGISTRY := change-it.dkr.ecr.us-west-2.amazonaws.com

.PHONY: ci
ci: deps deps_check lint build test

.PHONY: deps
deps:
	GOSUMDB=off GO111MODULE=on GOPROXY=direct go mod download
	GOSUMDB=off GO111MODULE=on GOPROXY=direct go mod vendor

.PHONY: deps_check
deps_check:
	@test -z "$(shell git status -s ./vendor ./Gopkg.*)"

.PHONY: grpcgen
grpcgen:
	protoc -I api api/service.proto --go_out=plugins=grpc:api

.PHONY: gqlgen
gqlgen:
	cd srvgql && \
	rm -f generated.go models/*_gen.go && \
	go run scripts/gqlgen.go -v

.PHONY: build
build:
	go build -o artifacts/svc

.PHONY: run
run:
	go run ./main.go

.PHONY: lint
lint:
	golangci-lint run

.PHONY: test
test:
	go test -cover -v `go list ./...`

.PHONY: test_integration
test_integration:
	INTEGRATION_TEST=YES go test -cover -v `go list ./...`

.PHONY: dockerise
dockerise:
	docker build -t ${IMAGE_NAME}:${IMAGE_TAG} -f Dockerfile .
	docker tag ${IMAGE_NAME}:${IMAGE_TAG} ${REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG}

.PHONY: deploy
deploy:
	`AWS_SHARED_CREDENTIALS_FILE=~/.aws/credentials AWS_PROFILE=xid aws ecr get-login --region us-west-2 --no-include-email`
	docker push ${REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG}
	#docker tag ${IMAGE_NAME}:${IMAGE_TAG} ${REGISTRY}/${IMAGE_NAME}:latest
	#docker push ${REGISTRY}/${IMAGE_NAME}:latest


.PHONY: mockgen
mockgen:
	mockgen -source=service/service.go -destination=service/mock/deps.go
	mockgen -source=srvhttp/service.go -destination=srvhttp/mock/service.go
	mockgen -source=srvgrpc/service.go -destination=srvgrpc/mock/service.go
	mockgen -source=srvgql/service.go -destination=srvgql/mock/service.go

.PHONY: run_postgresql
run_postgresql:
	docker run -d --name dummy_postgresql -e POSTGRES_DB=dummy -v ${ROOT_DIR}/tmp/sql/data:/var/lib/postgresql/data -p 5432:5432 postgres:11

.PHONY: run_redis
run_redis:
	docker run --name dummy_redis -p 6379:6379 -d redis

.PHONY: start_deps
start_deps:
	docker start dummy_redis
	docker start dummy_postgresql

.PHONY: stop_deps
stop_deps:
	docker stop dummy_redis
	docker stop dummy_postgresql

#.PHONY: exec_redis_sh
#exec_redis_sh:
#	docker exec -it dummy_redis sh
#    # redis-cli