package server

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"tages/config"
	pb "tages/internal/grpc_service/files.service"
	"tages/internal/service"

	"golang.org/x/sync/semaphore"

	"google.golang.org/protobuf/types/known/emptypb"
)

// Server -- структура сервера.
type Server struct {
	pb.UnimplementedFileServiceServer
	CreateLimit *semaphore.Weighted
	GetLimit    *semaphore.Weighted
	ListLimit   *semaphore.Weighted

	FileService service.IFileService

	Cfg *config.Config
}

// FileInfo -- структура файла.
type FileInfo struct {
	FilePath string
	buffer   *bytes.Buffer
	File     *os.File
}

// Create сохраняет файл.
func (s *Server) Create(stream pb.FileService_CreateServer) error {
	// семафором ограничиваем кол-во запросов.
	if err := s.CreateLimit.Acquire(stream.Context(), 1); err != nil {
		slog.Error("acquire createLimit sem")

		return errors.New("acquire createLimit sem")
	}
	defer s.CreateLimit.Release(1)

	file := FileInfo{}

	for {
		// получаем данные запроса.
		req, err := stream.Recv()
		if file.FilePath == "" {
			// создаем директорию для хранения созданных файлов.
			err := os.MkdirAll(s.Cfg.StoragePath, os.ModePerm)
			if err != nil {
				slog.Error("make file directory")

				return errors.New("make file directory")
			}

			file.FilePath = filepath.Join(s.Cfg.StoragePath, req.GetFilename())
			exists, err := s.FileService.CheckExistence(req.GetFilename())
			if err != nil {
				slog.Error("check existence while creating")

				return errors.New("check existence while creating")
			}

			if exists {
				return errors.New("name already exists")
			}

			dstFile, err := os.Create(file.FilePath)
			if err != nil {
				slog.Error("create file")

				return errors.New("create file")
			}

			file.File = dstFile
			defer file.File.Close()
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			slog.Error("make file")

			return errors.New("make file")
		}

		// пишем содержимое файла.
		buf := req.GetData()
		if _, err = file.File.Write(buf); err != nil {
			slog.Error("write to file")

			return errors.New("write to file")
		}
	}

	return stream.SendAndClose(&pb.CreateResponse{Filename: filepath.Base(file.FilePath)})
}

// List возвращает список файлов.
func (s *Server) List(ctx context.Context, _ *emptypb.Empty) (*pb.ListResponse, error) {
	// семафором ограничиваем кол-во запросов.
	if err := s.ListLimit.Acquire(ctx, 1); err != nil {
		slog.Error("acquire listLimit sem")

		return nil, errors.New("acquire listLimit sem")
	}
	defer s.ListLimit.Release(1)

	files, err := s.FileService.List()
	if err != nil {
		slog.Error("get list of files")

		return nil, errors.New("get list of files")
	}

	pbFileInfos := make([]*pb.FileInfo, 0, len(files))

	for i := range files {
		pbFileInfos = append(pbFileInfos, files[i].ToGRPC())
	}

	return &pb.ListResponse{Files: pbFileInfos}, nil
}

// Get получает загруженный файл.
func (s *Server) Get(req *pb.GetRequest, server pb.FileService_GetServer) error {
	// семафором ограничиваем кол-во запросов.
	if err := s.GetLimit.Acquire(server.Context(), 1); err != nil {
		slog.Error("acquire getLimit sem")

		return errors.New("acquire getLimit sem")
	}

	exists, err := s.FileService.CheckExistence(req.GetFilename())
	if err != nil {
		slog.Error("check existence while getting")

		return errors.New("check existence while getting")
	}

	if !exists {
		return errors.New("name does not exists")
	}

	filePath := filepath.Join(s.Cfg.StoragePath, req.Filename)
	file, err := os.Open(filePath)
	if err != nil {
		slog.Error("open file")

		return errors.New("open file")
	}

	defer func() {
		file.Close()
		s.GetLimit.Release(1)
	}()

	buf := make([]byte, 2048)
	for {
		n, err := file.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			slog.Error("read file")

			return errors.New("read file")
		}

		if err := server.Send(&pb.GetResponse{Data: buf[:n]}); err != nil {
			slog.Error("send file to server")

			return errors.New("send file to server")
		}
	}

	return nil
}
