# компилирует и запускает проект
run:
	protoc -I protos protos/file.proto --go_out=grpc_service --go-grpc_out=grpc_service; go run main.go -c local.yaml

# стартует генераторы проекта
gen:
	go generate ./...

# запускает все тесты проекта
test:
	go test ./... -cover

