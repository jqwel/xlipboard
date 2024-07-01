package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/quic-go/quic-go"
	quichttp "github.com/quic-go/quic-go/http3"
	"google.golang.org/protobuf/proto"

	"github.com/jqwel/xlipboard/src/rpc/iconst"
	"github.com/jqwel/xlipboard/src/rpc/service"
	"github.com/jqwel/xlipboard/src/utils"
)

var QClient *http.Client

func init() {
	roundTripper := &quichttp.RoundTripper{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		QuicConfig: &quic.Config{
			//KeepAlivePeriod: time.Second * 10,
			//MaxIncomingStreams:    1000,
			//MaxIncomingUniStreams: 1000,
			//Allow0RTT: true,
			//EnableDatagrams:       true,
		},
	}
	QClient = &http.Client{
		Transport: roundTripper,
	}
}

func GetConnQ(target string) func(route string, timeout time.Duration, input interface{}, output interface{}) (interface{}, error) {
	return func(route string, timeout time.Duration, input interface{}, output interface{}) (interface{}, error) {
		data, err := proto.Marshal(input.(proto.Message))
		if err != nil {
			return output, err
		}
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		url := fmt.Sprintf("https://%s%s", target, route)
		request, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(data))
		if err != nil {
			return output, err
		}
		request.Header.Set("Content-Type", "application/octet-stream")
		request.Header.Set("Authorization", "Bearer "+requestAuthkey)
		resp, err := QClient.Do(request)
		if err != nil {
			return output, err
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return output, err
		}
		if resp.StatusCode != http.StatusOK {
			return output, errors.New(strings.Replace(string(body), "\n", "", -1))
		}
		err = proto.Unmarshal(body, output.(proto.Message))
		if err != nil {
			return output, err
		}
		return output, nil
	}
}

func SayHelloQ(target string, currentChangeAt int64) (*service.HelloReply, error) {
	call := GetConnQ(target)
	reply, err := call(iconst.RoutePathSayHello, time.Millisecond*300, &service.HelloRequest{
		Timestamp: currentChangeAt,
		RequestID: utils.RandStringBytes(32),
	}, &service.HelloReply{})
	return ReturnCheck(reply.(*service.HelloReply), err)
}

func SayHowAreYouQ(target string, forChangeAt int64) (*service.HowAreYouReply, error) {
	call := GetConnQ(target)
	reply, err := call(iconst.RoutePathSayHowAreYou, time.Second*5, &service.HowAreYouRequest{
		Timestamp: forChangeAt,
	}, &service.HowAreYouReply{})
	return ReturnCheck(reply.(*service.HowAreYouReply), err)
}

func ReadDirQ(target string, path string) (*service.ReadDirReply, error) {
	call := GetConnQ(target)
	reply, err := call(iconst.RoutePathReadDir, time.Second*5, &service.ReadDirRequest{
		Path: path,
	}, &service.ReadDirReply{})
	return ReturnCheck(reply.(*service.ReadDirReply), err)
}

func OpenQ(target string, path string) (*service.OpenReply, error) {
	call := GetConnQ(target)
	reply, err := call(iconst.RoutePathOpen, time.Second*5, &service.OpenRequest{
		Path: path,
	}, &service.OpenReply{})
	return ReturnCheck(reply.(*service.OpenReply), err)
}

func ReleaseQ(target string, path string, fh uint64) (*service.ReleaseReply, error) {
	call := GetConnQ(target)
	reply, err := call(iconst.RoutePathRelease, time.Second*5, &service.ReleaseRequest{
		Path: path,
		Fh:   fh,
	}, &service.ReleaseReply{})
	return ReturnCheck(reply.(*service.ReleaseReply), err)
}

func ReadFileQ(target string, path string, buf []byte, offset int64, fh uint64, mul uint32) (*service.ReadFileReply, error) {
	call := GetConnQ(target)
	reply, err := call(iconst.RoutePathReadFile, time.Second*15, &service.ReadFileRequest{
		Path:   path,
		Offset: offset,
		LenBuf: uint32(len(buf)),
		Fh:     fh,
		Mul:    mul,
	}, &service.ReadFileReply{})
	return ReturnCheck(reply.(*service.ReadFileReply), err)
}

func StatQ(target string, path string) (*service.StatReply, error) {
	call := GetConnQ(target)
	reply, err := call(iconst.RoutePathStat, time.Second*5, &service.StatRequest{
		Path: path,
	}, &service.StatReply{})
	return ReturnCheck(reply.(*service.StatReply), err)
}
