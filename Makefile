# компилирует и запускает проект
run:
	protoc -I internal internal/protos/tages_test/fileservice/file.proto --go_out=internal/grpc_service --go-grpc_out=internal/grpc_service; go run main.go -c local.yaml

# стартует генераторы проекта
gen:
	go generate ./...

# запускает все тесты проекта
test:
	go test ./... -cover

