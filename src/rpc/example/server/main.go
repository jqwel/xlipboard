package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"

	"github.com/jqwel/xlipboard/src/utils/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	"github.com/jqwel/xlipboard/src/rpc/service"
	"github.com/jqwel/xlipboard/src/rpc/xlipboard/server"
	"github.com/jqwel/xlipboard/src/utils"
)

// authInterceptor 检查metadata中的authkey
func authInterceptor(ctx context.Context) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx, fmt.Errorf("no metadata")
	}
	v := md.Get("authkey")
	if len(v) != 1 || v[0] != "123456" {
		return ctx, fmt.Errorf("invalid authkey")
	}
	return ctx, nil
}

// unaryInterceptor 包装authInterceptor，使其符合grpc.UnaryServerInterceptor类型
func unaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// 在调用handler之前调用authInterceptor
	newCtx, err := authInterceptor(ctx)
	if err != nil {
		return nil, err
	}
	// 继续处理请求
	return handler(newCtx, req)
}

func main() {
	certificate, _, err := utils.GenerateCertificate(nil)
	if err != nil {
		logger.Logger.Errorln(err)
		return
	}
	cred := credentials.NewTLS(&tls.Config{
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{certificate},
	})

	listener, _ := net.Listen("tcp", ":9090")
	grpcServer := grpc.NewServer(grpc.Creds(cred), grpc.UnaryInterceptor(unaryInterceptor))
	service.RegisterSyncServiceServer(grpcServer, &server.Server{})
	err = grpcServer.Serve(listener)
	if err != nil {
		fmt.Println(err)
		return
	}
}
