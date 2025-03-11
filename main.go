package main

import (
	"flag"
	"log"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"tages/config"
	pb "tages/internal/grpc_service/files.service"
	"tages/internal/server"
	"tages/internal/service"

	"golang.org/x/sync/semaphore"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfgPath := flag.String("c", "config.yaml", "path to configuration file")
	flag.Parse()

	cfg, err := config.ParseConfig(*cfgPath)
	if err != nil {
		log.Fatal("parse config error: ", err)
	}

	if err = cfg.Validate(); err != nil {
		log.Fatal("config validation error: ", err)
	}

	listener, err := net.Listen("tcp", ":1037")
	if err != nil {
		slog.Error("run server")

		return
	}

	grpcServer := grpc.NewServer()
	pb.RegisterFileServiceServer(grpcServer, &server.Server{
		CreateLimit: semaphore.NewWeighted(int64(cfg.CreateLimit)),
		GetLimit:    semaphore.NewWeighted(int64(cfg.GetLimit)),
		ListLimit:   semaphore.NewWeighted(int64(cfg.ListLimit)),
		FileService: &service.FileService{Cfg: &cfg},
		Cfg:         &cfg,
	})
	reflection.Register(grpcServer)

	slog.Info("server is on 1037 port")

	if err := grpcServer.Serve(listener); err != nil {
		slog.Error("serve server")

		return
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop

	grpcServer.Stop()
	slog.Info("Gracefully stopped")
}
