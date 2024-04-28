package rpc_app

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"

	"google.golang.org/grpc/keepalive"

	"github.com/jqwel/xlipboard/src/rpc/iconst"
	"github.com/jqwel/xlipboard/src/utils/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	"github.com/jqwel/xlipboard/src/application"
	"github.com/jqwel/xlipboard/src/rpc/service"
	"github.com/jqwel/xlipboard/src/rpc/xlipboard/server"
	"github.com/jqwel/xlipboard/src/utils"
)

func RunGrpcServer(cert, priv string) error {
	certificate, _, err := utils.GenerateCertificate(&utils.CertificateData{
		Certificate: cert,
		PrivateKey:  priv,
	})
	if err != nil {
		logger.Logger.Errorln(err)
		return err
	}
	cred := credentials.NewTLS(&tls.Config{
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{certificate},
	})

	listener, err := net.Listen("tcp", ":"+application.App.Config.Port)
	if err != nil {
		fmt.Println(err)
		return err
	}
	opts := []grpc.ServerOption{
		grpc.Creds(cred),
		grpc.UnaryInterceptor(unaryInterceptor),
	}
	opts = append(opts, grpc.KeepaliveParams(keepalive.ServerParameters{
		//Time:    time.Second,
		//Timeout: time.Second * 3,
	}))
	//opts = append(opts, grpc.MaxSendMsgSize(1024*1024*128))
	opts = append(opts, grpc.MaxRecvMsgSize(iconst.MaxMsgSize))
	opts = append(opts, grpc.MaxSendMsgSize(iconst.MaxMsgSize))
	opts = append(opts, grpc.MaxConcurrentStreams(1024))

	grpcServer := grpc.NewServer(opts...)
	application.App.Server = grpcServer
	service.RegisterSyncServiceServer(grpcServer, server.ServerInstance)
	return grpcServer.Serve(listener)
}

// authInterceptor 检查metadata中的authkey
func authInterceptor(ctx context.Context) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx, fmt.Errorf("no metadata")
	}
	v := md.Get("authkey")
	if len(v) != 1 || v[0] != application.App.GetAuthKey() {
		return ctx, fmt.Errorf("invalid authkey")
	}
	return ctx, nil
}

// unaryInterceptor 包装authInterceptor，使其符合grpc.UnaryServerInterceptor类型
func unaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	newCtx, err := authInterceptor(ctx)
	if err != nil {
		return nil, err
	}
	return handler(newCtx, req)
}
