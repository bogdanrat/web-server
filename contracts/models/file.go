package models

import "time"

type GetFilesResponse struct {
	Key          string     `json:"key" csv:"Key"`
	LastModified *time.Time `json:"last_modified,omitempty" csv:"Last modified"`
	Size         uint64     `json:"size" csv:"Size"`
	StorageClass string     `json:"storage_class" csv:"-"`
}

type DeleteFileRequest struct {
	Key string `json:"key"`
}

type DeleteFilesRequest struct {
	Prefix *string `json:"prefix,omitempty"`
}

type GetFileRequest struct {
	FileName string `json:"file_name" form:"file_name"`
}
