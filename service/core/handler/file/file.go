package file

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/bogdanrat/web-server/contracts/models"
	"github.com/bogdanrat/web-server/contracts/proto/storage_service"
	"github.com/bogdanrat/web-server/service/core/config"
	"github.com/bogdanrat/web-server/service/core/lib"
	"github.com/bogdanrat/web-server/service/core/render"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"io"
	"mime/multipart"
	"net/http"
	"sync"
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

func (h *Handler) PostFiles(c *gin.Context) {
	if err := c.Request.ParseMultipartForm(100000); err != nil {
		jsonErr := models.NewInternalServerError("cannot parse multipart form")
		c.JSON(jsonErr.StatusCode, jsonErr)
		return
	}

	wg := sync.WaitGroup{}

	files := c.Request.MultipartForm.File["files"]
	for _, file := range files {
		wg.Add(1)
		go func(fileHeader *multipart.FileHeader) {
			defer wg.Done()

			jsonErr := h.uploadFile(fileHeader)
			if jsonErr != nil {
				c.JSON(jsonErr.StatusCode, jsonErr)
				return
			}
		}(file)
	}

	wg.Wait()

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

	fileName := file.Filename
	imagesPrefix := config.AppConfig.Services.Storage.ImagesPrefix
	if lib.IsImage(fileName) && imagesPrefix != "" {
		fileName = fmt.Sprintf("%s/%s", imagesPrefix, fileName)
	}

	request := &storage_service.UploadFileRequest{
		Data: &storage_service.UploadFileRequest_Info{
			Info: &storage_service.FileInfo{
				Size:     uint32(file.Size),
				FileName: fileName,
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

func (h *Handler) GetFile(c *gin.Context) {
	deadline := time.Now().Add(time.Millisecond * time.Duration(h.RPC.Deadline))
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	stream, err := h.RPC.Client.GetFiles(ctx, &storage_service.GetFilesRequest{})
	if err != nil {
		if jsonErr := lib.HandleRPCError(err); err != nil {
			c.JSON(jsonErr.StatusCode, jsonErr)
			return
		}
	}

	files, jsonErr := h.getAllFiles(stream)
	if jsonErr != nil {
		c.JSON(jsonErr.StatusCode, jsonErr)
		return
	}

	data := make(map[string]interface{})
	data["files"] = files

	_ = render.Template(c.Writer, c.Request, "file.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

func (h *Handler) PostFile(c *gin.Context) {
	request := &models.GetFileRequest{}
	if err := c.ShouldBindJSON(request); err != nil {
		jsonErr := models.NewBadRequestError("file name is required", "file_name")
		c.JSON(jsonErr.StatusCode, jsonErr)
		return
	}

	fileName := request.FileName

	deadline := time.Now().Add(time.Millisecond * time.Duration(h.RPC.Deadline))
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	stream, err := h.RPC.Client.GetFile(ctx, &storage_service.GetFileRequest{
		FileName: fileName,
	})

	if err != nil {
		if jsonErr := lib.HandleRPCError(err); err != nil {
			c.JSON(jsonErr.StatusCode, jsonErr)
			return
		}
	}

	fileData := &bytes.Buffer{}

	for {
		req, err := stream.Recv()
		if err != nil {
			// no more data
			if err == io.EOF {
				break
			}
			if jsonErr := lib.HandleRPCError(err); err != nil {
				c.JSON(jsonErr.StatusCode, jsonErr)
				c.Status(jsonErr.StatusCode)
				return
			}
		}

		chunk := req.GetChunkData()
		_, err = fileData.Write(chunk)

		if err != nil {
			jsonErr := lib.HandleRPCError(err)
			c.JSON(jsonErr.StatusCode, jsonErr)
			return
		}
	}

	c.Header("Content-Type", "image/jpeg")
	c.Header("Access-Control-Expose-Headers", "Content-Disposition")
	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Status(http.StatusOK)
	c.Writer.Write(fileData.Bytes())
}

func (h *Handler) GetFiles(c *gin.Context) {
	deadline := time.Now().Add(time.Millisecond * time.Duration(h.RPC.Deadline))
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	stream, err := h.RPC.Client.GetFiles(ctx, &storage_service.GetFilesRequest{})
	if err != nil {
		if jsonErr := lib.HandleRPCError(err); err != nil {
			c.JSON(jsonErr.StatusCode, jsonErr)
			return
		}
	}

	files, jsonErr := h.getAllFiles(stream)
	if jsonErr != nil {
		c.JSON(jsonErr.StatusCode, jsonErr)
		return
	}

	c.JSON(http.StatusOK, files)
}

func (h *Handler) getAllFiles(stream storage_service.Storage_GetFilesClient) ([]*models.GetFilesResponse, *models.JSONError) {
	files := make([]*models.GetFilesResponse, 0)

	for {
		response, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			if jsonErr := lib.HandleRPCError(err); jsonErr != nil {
				return nil, jsonErr
			}
		}

		file := &models.GetFilesResponse{
			Key:          response.Object.GetKey(),
			Size:         response.Object.GetSize(),
			StorageClass: response.Object.GetStorageClass(),
		}
		lastModified, err := time.Parse(time.RFC3339, response.Object.GetLastModified())
		if err == nil && !lastModified.IsZero() {
			file.LastModified = &lastModified
		}

		files = append(files, file)
	}

	return files, nil
}

func (h *Handler) DeleteFile(c *gin.Context) {
	request := &models.DeleteFileRequest{}

	if err := c.ShouldBindJSON(request); err != nil {
		jsonErr := models.NewBadRequestError("object key is required", "key")
		c.JSON(jsonErr.StatusCode, jsonErr)
		return
	}

	deadline := time.Now().Add(time.Millisecond * time.Duration(h.RPC.Deadline))
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	_, err := h.RPC.Client.DeleteFile(ctx, &storage_service.DeleteFileRequest{
		Key: request.Key,
	})
	if err != nil {
		if jsonErr := lib.HandleRPCError(err); err != nil {
			c.JSON(jsonErr.StatusCode, jsonErr)
			return
		}
	}

	c.Status(http.StatusOK)
}

func (h *Handler) DeleteFiles(c *gin.Context) {
	deadline := time.Now().Add(time.Millisecond * time.Duration(h.RPC.Deadline))
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	_, err := h.RPC.Client.DeleteFiles(ctx, &storage_service.DeleteFilesRequest{})
	if err != nil {
		if jsonErr := lib.HandleRPCError(err); err != nil {
			c.JSON(jsonErr.StatusCode, jsonErr)
			return
		}
	}

	c.Status(http.StatusOK)
}
