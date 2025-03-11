#set -ex

protoc -I protos protos/file.proto --go_out=grpc_service --go-grpc_out=grpc_service
