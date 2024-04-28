package main

import (
	"context"
	"crypto/tls"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/jqwel/xlipboard/src/utils/logger"

	"github.com/jqwel/xlipboard/src/rpc/service"
	"github.com/jqwel/xlipboard/src/utils"
)

type ClientTokenAuth struct {
}

func (c *ClientTokenAuth) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	return map[string]string{
		"authkey": "123456",
	}, nil
}
func (c *ClientTokenAuth) RequireTransportSecurity() bool {
	return true
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
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(cred))
	opts = append(opts, grpc.WithPerRPCCredentials(&ClientTokenAuth{}))
	conn, err := grpc.Dial("localhost:9090", opts...)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	client := service.NewSyncServiceClient(conn)
	hello, err := client.Ping(context.Background(), &service.PingRequest{
		Msg: "ping",
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(hello.GetMsg())
	//conn.Close()
}
