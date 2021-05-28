package file

import (
	"bufio"
	"context"
	"fmt"
	"github.com/bogdanrat/web-server/contracts/models"
	"github.com/bogdanrat/web-server/contracts/proto/storage_service"
	"github.com/bogdanrat/web-server/service/core/lib"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

type RPCConfig struct {
	Client      storage_service.StorageClient
	CallOptions []grpc.CallOption
	Deadline    int64
}

type Handler struct {
	RPC *RPCConfig
}

func NewHandler(rpcConfig *RPCConfig) *Handler {
	return &Handler{
		RPC: rpcConfig,
	}
}

func (h *Handler) PostFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		jsonErr := models.NewBadRequestError("missing file", "file")
		c.JSON(jsonErr.StatusCode, jsonErr.Description)
		return
	}

	jsonErr := h.uploadFile(file)
	if jsonErr != nil {
		c.JSON(jsonErr.StatusCode, jsonErr.Description)
		return
	}

	c.Status(http.StatusCreated)
}

func (h *Handler) uploadFile(file *multipart.FileHeader) *models.JSONError {
	fileHeader, err := file.Open()
	if err != nil {
		return models.NewInternalServerError("cannot open file", "file")
	}

	deadline := time.Now().Add(time.Millisecond * time.Duration(h.RPC.Deadline))
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	uploadStream, err := h.RPC.Client.UploadFile(ctx)
	if err != nil {
		return models.NewInternalServerError("cannot open upload stream")
	}

	request := &storage_service.UploadFileRequest{
		Data: &storage_service.UploadFileRequest_Info{
			Info: &storage_service.FileInfo{
				Size:     uint32(file.Size),
				FileName: file.Filename,
			},
		},
	}

	if err = uploadStream.Send(request); err != nil {
		return lib.HandleRPCError(err)
	}

	reader := bufio.NewReader(fileHeader)
	// send file in chunks of 1 KB
	buffer := make([]byte, 1024)

	for {
		n, err := reader.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			return models.NewInternalServerError(fmt.Sprintf("cannot read file chunk: %s", err.Error()))
		}

		request := &storage_service.UploadFileRequest{
			Data: &storage_service.UploadFileRequest_ChunkData{
				ChunkData: buffer[:n],
			},
		}

		err = uploadStream.Send(request)
		if err != nil {
			return lib.HandleRPCError(err)
		}
	}

	response, err := uploadStream.CloseAndRecv()
	if err != nil {
		return models.NewInternalServerError(fmt.Sprintf("cannot receive response: %s", err.Error()))
	}
	if response == nil {
		return models.NewInternalServerError("received empty response")
	}

	return nil
}
