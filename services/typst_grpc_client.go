package services

import (
	"context"
	"fmt"

	"github.com/goquizvibe/config"
	"github.com/goquizvibe/internal/grpc/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type TypstGRPCClient struct {
	conn   *grpc.ClientConn
	client proto.TypstCompilerClient
}

func NewTypstGRPCClient(cfg *config.Config) (*TypstGRPCClient, error) {
	addr := cfg.ServiceConfig.TypstServiceAddr
	if addr == "" {
		addr = "localhost:9090"
	}

	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("grpc dial: %w", err)
	}

	return &TypstGRPCClient{
		conn:   conn,
		client: proto.NewTypstCompilerClient(conn),
	}, nil
}

func (c *TypstGRPCClient) Compile(ctx context.Context, req *proto.CompileRequest) (*proto.CompileResponse, error) {
	resp, err := c.client.Compile(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("grpc compile: %w", err)
	}
	return resp, nil
}

func (c *TypstGRPCClient) Close() error {
	return c.conn.Close()
}
