package models

import "time"

type GetFilesResponse struct {
	Key          string     `json:"key"`
	LastModified *time.Time `json:"last_modified,omitempty"`
	Size         uint64     `json:"size"`
	StorageClass string     `json:"storage_class"`
}
