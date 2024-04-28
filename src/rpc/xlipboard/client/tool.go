package client

import (
	"context"
	"crypto/tls"
	"errors"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"

	"github.com/jqwel/xlipboard/src/rpc/iconst"
	"github.com/jqwel/xlipboard/src/utils"
	"github.com/jqwel/xlipboard/src/utils/logger"
)

var requestAuthkey = ""
var valueCred credentials.TransportCredentials

var muInitSettings sync.Mutex

func InitSettings(authkey string, cert, priv string) {
	muInitSettings.Lock()
	defer muInitSettings.Unlock()
	if requestAuthkey != "" || authkey == "" {
		return
	}
	requestAuthkey = authkey

	certificate, _, err := utils.GenerateCertificate(&utils.CertificateData{
		Certificate: cert,
		PrivateKey:  priv,
	})
	if err != nil {
		logger.Logger.Errorln(err)
	}
	valueCred = credentials.NewTLS(&tls.Config{
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{certificate},
	})
}

var mu sync.Mutex
var sm = sync.Map{}

func GetConn(target string) (*grpc.ClientConn, error) {
	mu.Lock()
	defer mu.Unlock()
	var fnCreateConn = func(target string) (*grpc.ClientConn, error) {
		var opts []grpc.DialOption
		opts = append(opts, grpc.WithTransportCredentials(valueCred))
		opts = append(opts, grpc.WithPerRPCCredentials(&ClientTokenAuth{}))
		opts = append(opts, grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(iconst.MaxMsgSize),
			grpc.MaxCallSendMsgSize(iconst.MaxMsgSize),
		))
		opts = append(opts, grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                time.Second * 10,
			Timeout:             time.Second * 3,
			PermitWithoutStream: false,
		}))
		return grpc.Dial(target, opts...)
	}

	var connMap *ConnMap
	value, ok := sm.Load(target)
	if !ok {
		connMap = NewConnMap(target, fnCreateConn)
		sm.Store(target, connMap)
	} else {
		connMap = value.(*ConnMap)
	}
	conn, b, err := connMap.GetOne()
	if err != nil {
		return nil, err
	}
	for !b {
		time.Sleep(time.Millisecond * 200)
		conn, b, err = connMap.GetOne()
		if err != nil {
			return nil, err
		}
	}
	return conn, nil
}
func ReleaseConn(target string, conn *grpc.ClientConn, e error) error {
	var connMap *ConnMap
	value, ok := sm.Load(target)
	if !ok {
		return errors.New("not found")
	}
	connMap = value.(*ConnMap)
	connMap.PutBack(conn, e)
	return nil
}

type ClientTokenAuth struct {
}

func (c *ClientTokenAuth) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	return map[string]string{
		"authkey": requestAuthkey,
	}, nil
}
func (c *ClientTokenAuth) RequireTransportSecurity() bool {
	return true
}

func ReturnCheck[T interface {
	GetMessage() string
}](reply T, err error) (T, error) {
	if err != nil {
		if status.Code(err) == codes.DeadlineExceeded {
			return reply, errors.New(iconst.Timeout)
		}
		return reply, err
	}
	if reply.GetMessage() != iconst.Success {
		return reply, errors.New("not success")
	}
	return reply, nil
}
