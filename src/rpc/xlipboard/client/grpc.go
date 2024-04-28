package client

import (
	"context"
	"io"
	"time"

	"github.com/jqwel/xlipboard/src/rpc/service"
	"github.com/jqwel/xlipboard/src/utils"
)

func SayHelloG(target string, currentChangeAt int64) (*service.HelloReply, error) {
	conn, err := GetConn(target)
	if err != nil {
		return nil, err
	}
	defer func() { ReleaseConn(target, conn, err) }()

	client := service.NewSyncServiceClient(conn)
	requestID := utils.RandStringBytes(32)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	reply, err := client.SayHello(ctx, &service.HelloRequest{
		Timestamp: currentChangeAt,
		RequestID: requestID,
	})
	return ReturnCheck(reply, err)
}

func SayHowAreYouG(target string, forChangeAt int64) (*service.HowAreYouReply, error) {
	conn, err := GetConn(target)
	if err != nil {
		return nil, err
	}
	defer func() { ReleaseConn(target, conn, err) }()

	client := service.NewSyncServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	reply, err := client.SayHowAreYou(ctx, &service.HowAreYouRequest{
		Timestamp: forChangeAt,
	})
	return ReturnCheck(reply, err)
}

func ReadDirG(target string, path string) (*service.ReadDirReply, error) {
	conn, err := GetConn(target)
	if err != nil {
		return nil, err
	}
	defer func() { ReleaseConn(target, conn, err) }()

	client := service.NewSyncServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	reply, err := client.ReadDir(ctx, &service.ReadDirRequest{
		Path: path,
	})
	return ReturnCheck(reply, err)
}

func OpenG(target string, path string) (*service.OpenReply, error) {
	conn, err := GetConn(target)
	if err != nil {
		return nil, err
	}
	defer func() { ReleaseConn(target, conn, err) }()

	client := service.NewSyncServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	reply, err := client.Open(ctx, &service.OpenRequest{
		Path: path,
	})
	return ReturnCheck(reply, err)
}

func ReleaseG(target string, path string, fh uint64) (*service.ReleaseReply, error) {
	conn, err := GetConn(target)
	if err != nil {
		return nil, err
	}
	defer func() { ReleaseConn(target, conn, err) }()

	client := service.NewSyncServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	reply, err := client.Release(ctx, &service.ReleaseRequest{
		Path: path,
		Fh:   fh,
	})
	return ReturnCheck(reply, err)
}

func ReadFileG(target string, path string, buf []byte, offset int64, fh uint64, mul uint32) (*service.ReadFileReply, error) {
	conn, err := GetConn(target)
	if err != nil {
		return nil, err
	}
	defer func() { ReleaseConn(target, conn, err) }()

	client := service.NewSyncServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	reply, err := client.ReadFile(ctx, &service.ReadFileRequest{
		Path:   path,
		Offset: offset,
		LenBuf: uint32(len(buf)),
		Fh:     fh,
		Mul:    mul,
	})
	return ReturnCheck(reply, err)
}

func ReadFileStreamG(target string, path string, buf []byte, offset int64, fh uint64, mul uint32) (*service.ReadFileReply, error) {
	conn, err := GetConn(target)
	if err != nil {
		return nil, err
	}
	defer func() { ReleaseConn(target, conn, err) }()
	client := service.NewSyncServiceClient(conn)
	lenBuf := uint32(len(buf))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	stream, err := client.ReadFileStream(ctx, &service.ReadFileRequest{
		Path:   path,
		Offset: offset,
		LenBuf: lenBuf,
		Fh:     fh,
		Mul:    mul,
	})
	if err != nil {
		return nil, err
	}
	reply0 := &service.ReadFileReply{
		Read: 0,
		EOF:  false,
	}
	for {
		reply, err := stream.Recv()
		if err == io.EOF {
			break
		}
		reply, err = ReturnCheck(reply, err)
		if err != nil {
			return nil, err
		}
		if reply.GetEOF() {
			break
		}
		reply0.Buf = append(reply0.Buf, reply.GetBuf()...)
		reply0.Read += reply.GetRead()
	}

	return reply0, nil
}

func StatG(target string, path string) (*service.StatReply, error) {
	conn, err := GetConn(target)
	if err != nil {
		return nil, err
	}
	defer func() { ReleaseConn(target, conn, err) }()
	client := service.NewSyncServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	reply, err := client.Stat(ctx, &service.StatRequest{
		Path: path,
	})
	return ReturnCheck(reply, err)
}
