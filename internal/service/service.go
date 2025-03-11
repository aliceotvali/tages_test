package service

import (
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"tages/config"
	"tages/internal/model"
	"time"
)

// IFileService - интерфейс сервиса.
type IFileService interface {
	CheckExistence(fileName string) (bool, error)
	List() ([]model.FileInfo, error)
}

// FileService -- структура сервиса.
type FileService struct {
	Cfg *config.Config
}

// List возвращает список файлов.
func (s *FileService) List() ([]model.FileInfo, error) {
	// считываем файлы из директории.
	files, err := os.ReadDir(s.Cfg.StoragePath)
	if err != nil {
		slog.Error("read files directory")

		return nil, errors.New("read files directory")
	}

	var fileInfos []model.FileInfo
	for _, f := range files {
		info, err := f.Info()
		if err != nil {
			slog.Error("get file info")

			return nil, errors.New("get file info")
		}

		fileInfos = append(fileInfos, model.FileInfo{
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
