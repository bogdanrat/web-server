package models

import "time"

type GetFilesResponse struct {
	Key          string     `json:"key"`
	LastModified *time.Time `json:"last_modified,omitempty"`
	Size         uint64     `json:"size"`
	StorageClass string     `json:"storage_class"`
}

type DeleteFileRequest struct {
	Key string `json:"key"`
}

type DeleteFilesRequest struct {
	Prefix string `json:"prefix,omitempty"`
}

type GetFileRequest struct {
	FileName string `json:"file_name" form:"file_name"`
}
