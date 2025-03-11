package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"sync"

	pb "tages/grpc_service/files.service"

	"time"

	"golang.org/x/sync/semaphore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	createLimit = semaphore.NewWeighted(10)  // 10 загрузок
	getLimit    = semaphore.NewWeighted(10)  // 10 скачиваний
	listLimit   = semaphore.NewWeighted(100) // 100 запросов списка файлов
	mu          sync.Mutex
)

const storagePath = "binary_files"

type FileServiceServer struct {
	pb.UnimplementedFileServiceServer
}

// FileInfo -- структура файла.
type FileInfo struct {
	FilePath string
	buffer   *bytes.Buffer
	File     *os.File
}

// Create сохраняет файл.
func (g *FileServiceServer) Create(stream pb.FileService_CreateServer) error {
	// семафором ограничиваем кол-во запросов.
	if err := createLimit.Acquire(stream.Context(), 1); err != nil {
		slog.Error("acquire createLimit sem")

		return errors.New("acquire createLimit sem")
	}
	defer createLimit.Release(1)

	file := FileInfo{}

	for {
		// получаем данные запроса.
		req, err := stream.Recv()
		if file.FilePath == "" {
			// создаем директорию для хранения созданных файлов.
			err := os.MkdirAll(storagePath, os.ModePerm)
			if err != nil {
				slog.Error("make file directory")

				return errors.New("make file directory")
			}

			file.FilePath = filepath.Join(storagePath, req.GetFilename())
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
	fileName := filepath.Base(file.FilePath)
	return stream.SendAndClose(&pb.CreateResponse{Filename: fileName})
}

// List возвращает список файлов.
func (s *FileServiceServer) List(ctx context.Context, _ *emptypb.Empty) (*pb.ListResponse, error) {
	// семафором ограничиваем кол-во запросов.
	if err := listLimit.Acquire(ctx, 1); err != nil {
		slog.Error("acquire listLimit sem")

		return nil, errors.New("acquire listLimit sem")
	}
	defer listLimit.Release(1)

	// считываем файлы из директории.
	files, err := os.ReadDir(storagePath)
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

	return &pb.ListResponse{Files: fileInfos}, nil
}

// Get получает загруженный файл.
func (s *FileServiceServer) Get(req *pb.GetRequest, server pb.FileService_GetServer) error {
	// семафором ограничиваем кол-во запросов.
	if err := getLimit.Acquire(server.Context(), 1); err != nil {
		slog.Error("acquire getLimit sem")

		return errors.New("acquire getLimit sem")
	}

	filePath := filepath.Join(storagePath, req.Filename)
	file, err := os.Open(filePath)
	if err != nil {
		slog.Error("open file")

		return errors.New("open file")
	}

	defer func() {
		file.Close()
		getLimit.Release(1)
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

func main() {
	listener, err := net.Listen("tcp", ":1037")
	if err != nil {
		slog.Error("run server: %w", err)

		return
	}

	grpcServer := grpc.NewServer()
	pb.RegisterFileServiceServer(grpcServer, &FileServiceServer{})
	reflection.Register(grpcServer)

	slog.Info("server is on 1037 port")

	if err := grpcServer.Serve(listener); err != nil {
		slog.Error("serve server: %w", err)

		return
	}
}
