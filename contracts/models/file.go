package models

import "time"

type GetFilesResponse struct {
	Key          string     `json:"key" csv:"Key" excel:"Key"`
	LastModified *time.Time `json:"last_modified,omitempty" csv:"Last modified" excel:"Last modified"`
	Size         uint64     `json:"size" csv:"Size" excel:"Size"`
	StorageClass string     `json:"storage_class" csv:"Storage Class" excel:"Storage Class"`
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
