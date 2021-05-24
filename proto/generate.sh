protoc -I ./auth_service ./auth_service/auth_service.proto --go_out=plugins=grpc:./auth_service
protoc -I ./auth_service ./auth_service/auth_service.proto --go_out=plugins=grpc:../service/auth/proto
protoc -I ./auth_service ./auth_service/auth_service.proto --go_out=plugins=grpc:../service/core/proto