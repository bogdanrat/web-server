package handler

import (
	"bytes"
	"context"
	"fmt"
	pb "github.com/bogdanrat/web-server/contracts/proto/storage_service"
	"github.com/bogdanrat/web-server/service/storage/config"
	"github.com/bogdanrat/web-server/service/storage/lib"
	"github.com/bogdanrat/web-server/service/storage/persistence"
	epb "google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"log"
)

type StorageServer struct {
	Storage *persistence.Storage
}

func New(storage *persistence.Storage) *StorageServer {
	return &StorageServer{
		Storage: storage,
	}
}

func (s *StorageServer) UploadFile(stream pb.Storage_UploadFileServer) error {
	req, err := stream.Recv()
	if err != nil {
		return logError(status.Errorf(codes.Unknown, "cannot receive file info: %s", err.Error()))
	}

	fileSize := req.GetInfo().GetSize()
	if fileSize > config.AppConfig.Upload.MaxFileSize {
		errorStatus := status.New(codes.ResourceExhausted, "invalid file size")
		details, err := errorStatus.WithDetails(&epb.BadRequest_FieldViolation{
			Field:       "size",
			Description: fmt.Sprintf("file size %s exceeds maximum size %s", lib.FormatSize(int(fileSize), 2), lib.FormatSize(int(config.AppConfig.Upload.MaxFileSize), 2)),
		})
		if err != nil {
			return logError(errorStatus.Err())
		}
		return logError(details.Err())
	}

	fileName := req.GetInfo().GetFileName()
	log.Printf("Request to upload %s\n", fileName)

	fileData := bytes.Buffer{}

	for {
		if err := contextError(stream.Context()); err != nil {
			return logError(err)
		}

		req, err := stream.Recv()
		if err != nil {
			// no more data
			if err == io.EOF {
				break
			}
			return logError(status.Errorf(codes.Unknown, "cannot receive chunk data: %v", err))
		}

		chunk := req.GetChunkData()
		_, err = fileData.Write(chunk)

		if err != nil {
			return logError(status.Errorf(codes.Internal, "cannot write chunk data: %v", err))
		}
	}

	reader := bytes.NewReader(fileData.Bytes())
	_, err = s.Storage.UploadFile(fileName, reader)

	if err != nil {
		return logError(status.Errorf(codes.Internal, "cannot upload file: %v", err))
	}
	if err := contextError(stream.Context()); err != nil {
		return logError(err)
	}

	err = stream.SendAndClose(&pb.UploadFileResponse{})
	if err != nil {
		return logError(status.Errorf(codes.Internal, "cannot send response: %v", err))
	}

	log.Printf("Uploaded image %s, size: %d", fileName, fileSize)
	return nil
}

func (s *StorageServer) GetFiles(req *pb.GetFilesRequest, stream pb.Storage_GetFilesServer) error {
	objects, err := s.Storage.GetFiles()
	if err != nil {
		return logError(status.Errorf(codes.Internal, "cannot get files: %v", err))
	}

	for _, object := range objects {
		if err = contextError(stream.Context()); err != nil {
			return logError(err)
		}

		err = stream.Send(&pb.GetFilesResponse{
			Object: object,
		})
		if err != nil {
			return logError(status.Errorf(codes.DataLoss, "cannot send file: %v", err))
		}
	}

	return nil
}

func (s *StorageServer) DeleteFile(ctx context.Context, req *pb.DeleteFileRequest) (*pb.DeleteFileResponse, error) {
	if err := s.Storage.DeleteFile(req.Key); err != nil {
		return nil, logError(status.Errorf(codes.Internal, "cannot delete: %v", err))
	}

	if err := contextError(ctx); err != nil {
		return nil, logError(err)
	}

	return &pb.DeleteFileResponse{}, nil
}

func (s *StorageServer) DeleteFiles(ctx context.Context, req *pb.DeleteFilesRequest) (*pb.DeleteFilesResponse, error) {
	if err := s.Storage.DeleteFiles(); err != nil {
		return nil, logError(status.Errorf(codes.Internal, "cannot delete: %v", err))
	}

	if err := contextError(ctx); err != nil {
		return nil, logError(err)
	}

	return &pb.DeleteFilesResponse{}, nil
}

func contextError(ctx context.Context) error {
	switch ctx.Err() {
	case context.Canceled:
		return status.Errorf(codes.Canceled, "request was canceled by the client")
	case context.DeadlineExceeded:
		return status.Errorf(codes.DeadlineExceeded, "request deadline was exceeded")
	}
	return nil
}

func logError(err error) error {
	if err != nil {
		log.Println(err)
	}
	return err
}
