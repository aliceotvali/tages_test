package service

import (
	"bytes"
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"tages/config"
	pb "tages/grpc_service/files.service"
	"time"
)

// IFileService - интерфейс сервиса.
type IFileService interface {
	CheckExistence(fileName string) (bool, error)
	List() ([]*pb.FileInfo, error)
}

// FileService -- структура сервиса.
type FileService struct {
	Cfg *config.Config
}

// FileInfo -- структура файла.
type FileInfo struct {
	FilePath string
	buffer   *bytes.Buffer
	File     *os.File
}

// List возвращает список файлов.
func (s *FileService) List() ([]*pb.FileInfo, error) {
	// считываем файлы из директории.
	files, err := os.ReadDir(s.Cfg.StoragePath)
	if err != nil {
		slog.Error("read files directory")

		return nil, errors.New("read files directory")
	}

	var fileInfos []*pb.FileInfo
	for _, f := range files {
		info, err := f.Info()
		if err != nil {
			slog.Error("get file info")

			return nil, errors.New("get file info")
		}

		fileInfos = append(fileInfos, &pb.FileInfo{
			Name:      f.Name(),
			CreatedAt: info.ModTime().Format(time.RFC3339),
			UpdatedAt: info.ModTime().Format(time.RFC3339),
		})
	}

	return fileInfos, nil
}

// CheckExistence проверят существование файла с переданным именем.
func (s *FileService) CheckExistence(fileName string) (bool, error) {
	_, err := os.Stat(filepath.Join(s.Cfg.StoragePath, fileName))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		}

		slog.Error("get file info by name")

		return false, errors.New("get fileInfo by name")
	}

	return true, nil
}
