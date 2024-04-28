package main

import (
	"context"
	"crypto/tls"
	"fmt"

	quic "github.com/quic-go/quic-go"
)

const addr = "localhost:3216"

func main() {
	err := clientMain()
	if err != nil {
		fmt.Println(err)
	}
}

func clientMain() error {
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"quic-echo-example"},
	}

	ctx := context.Background()
	session, err := quic.DialAddr(ctx, addr, tlsConf, nil)
	if err != nil {
		return err
	}

	stream, err := session.OpenStreamSync(ctx)
	if err != nil {
		return err
	}

	_, err = stream.Write([]byte("Hello, World!\n"))
	if err != nil {
		return err
	}

	buf := make([]byte, 100)
	n, err := stream.Read(buf)
	if err != nil {
		return err
	}

	fmt.Printf("Server's reply: %s\n", buf[:n])

	return nil
}
