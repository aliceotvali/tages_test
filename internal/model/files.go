package model

import (
	"bytes"
	"os"
	pb "tages/internal/grpc_service/files.service"
)

type (
	// FileStruct -- структура файла.
	FileStruct struct {
		FilePath string
		buffer   *bytes.Buffer
		File     *os.File
	}

	// FileInfo -- структура информации о файле.
	FileInfo struct {
		Name      string
		CreatedAt string
		UpdatedAt string
	}
)

// ToGRPC готовит FileInfo для передачи grpc.
func (f FileInfo) ToGRPC() *pb.FileInfo {
	return &pb.FileInfo{
		Name:      f.Name,
		CreatedAt: f.CreatedAt,
		UpdatedAt: f.UpdatedAt,
	}
}
