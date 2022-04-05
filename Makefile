PROJECT_NAME=web-crawler

AMQP_WORKER_PATH="cmd/amqp/main.go"
GRPC_SERVER_PATH="cmd/grpc/main.go"
CLIENT_PATH="cmd/client/main.go"

build:
	go build -o $(PROJECT_NAME)-api $(GRPC_SERVER_PATH)
	go build -o $(PROJECT_NAME)-worker $(AMQP_WORKER_PATH)
	go build -o $(PROJECT_NAME)-client $(CLIENT_PATH)

test:
	go test -short ./...

test-integration-acceptance:
	./tests/run.sh