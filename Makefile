.PHONY: release docker proto
.DEFAULT_GOAL := docker

proto:
	scripts/proto.sh

docker:
	docker build -t thrawn01/grpc-http-1:latest .

