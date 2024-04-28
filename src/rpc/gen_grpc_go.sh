#protoc --go_out=. --go_opt=paths=../service --go-grpc_out=. --go-grpc_opt=paths=../service proto/hello.proto
mkdir -p ./service
protoc --go_out=./service proto/sync.proto
protoc --go-grpc_out=./service proto/sync.proto
sleep 1

