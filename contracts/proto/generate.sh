protoc -I ./auth_service ./auth_service/auth_service.proto --go_out=plugins=grpc:./auth_service
protoc -I ./storage_service ./storage_service/storage_service.proto --go_out=plugins=grpc:./storage_service
