package main

import (
	"log"
	"net"
	"os"

	"github.com/goquizvibe/microservices/typst/internal/server"
	"github.com/goquizvibe/microservices/typst/internal/server/proto"
	"github.com/goquizvibe/microservices/typst/internal/service"
	"github.com/goquizvibe/pkg/storage"
	"google.golang.org/grpc"
)

func main() {
	cfg := storage.MinioConfig{
		Endpoint:  os.Getenv("MINIO_ENDPOINT"),
		AccessKey: os.Getenv("MINIO_ROOT_USER"),
		SecretKey: os.Getenv("MINIO_ROOT_PASSWORD"),
		Bucket:    os.Getenv("MINIO_BUCKET"),
	}
	if cfg.Endpoint == "" {
		cfg.Endpoint = "localhost:9000"
	}
	if cfg.Bucket == "" {
		cfg.Bucket = "goquizvibe"
	}

	minioClient, err := storage.NewMinioClient(cfg)
	if err != nil {
		log.Fatalf("failed to create minio client: %v", err)
	}

	compiler := service.NewCompiler()
	typstServer := server.NewServer(compiler, minioClient)

	grpcServer := grpc.NewServer()
	proto.RegisterTypstCompilerServer(grpcServer, typstServer)

	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "9090"
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Printf("typst service listening on :%s", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}