package rpc_app

import (
	"crypto/tls"
	"io"
	"net/http"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"google.golang.org/protobuf/proto"

	"github.com/jqwel/xlipboard/src/application"
	"github.com/jqwel/xlipboard/src/rpc/iconst"
	"github.com/jqwel/xlipboard/src/rpc/service"
	gs "github.com/jqwel/xlipboard/src/rpc/xlipboard/server"
	"github.com/jqwel/xlipboard/src/utils"
	"github.com/jqwel/xlipboard/src/utils/logger"
)

func RunQuicServer(cert, priv string) error {
	certificate, _, err := utils.GenerateCertificate(&utils.CertificateData{
		Certificate: cert,
		PrivateKey:  priv,
	})
	if err != nil {
		logger.Logger.Errorln(err)
		return err
	}
	handler := setupHandler()
	server := http3.Server{
		Handler: handler,
		Addr:    ":" + application.App.Config.Port,
		QuicConfig: &quic.Config{
			KeepAlivePeriod:       time.Second * 10,
			MaxIncomingStreams:    1000,
			MaxIncomingUniStreams: 1000,
			Allow0RTT:             true,
			EnableDatagrams:       true,
			//Tracer: qlog.DefaultTracer,
		},
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
			Certificates:       []tls.Certificate{certificate},
		},
	}
	return server.ListenAndServe()
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

// 定义一个中间件函数，用于验证请求头部信息
func headerValidator(next http.Handler) http.Handler {
	authkey := application.App.GetAuthKey()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				if err := err.(error); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
			}
		}()
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer "+authkey {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func setupHandler() http.Handler {
	mux := http.NewServeMux()
	var gs = gs.ServerInstance

	var receiveReq = func(r *http.Request) []byte {
		body, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		checkError(err)
		return body
	}
	var replyBack = func(w http.ResponseWriter, reply proto.Message, err error) {
		checkError(err)
		bs, err := proto.Marshal(reply)
		checkError(err)
		w.Write(bs)
	}

	mux.Handle(iconst.RoutePathSayHello, headerValidator(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req service.HelloRequest
		checkError(proto.Unmarshal(receiveReq(r), &req))
		reply, err := gs.SayHello(r.Context(), &req)
		replyBack(w, reply, err)
	})))

	mux.Handle(iconst.RoutePathSayHowAreYou, headerValidator(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req service.HowAreYouRequest
		checkError(proto.Unmarshal(receiveReq(r), &req))
		reply, err := gs.SayHowAreYou(r.Context(), &req)
		replyBack(w, reply, err)
	})))

	mux.Handle(iconst.RoutePathReadDir, headerValidator(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req service.ReadDirRequest
		checkError(proto.Unmarshal(receiveReq(r), &req))
		reply, err := gs.ReadDir(r.Context(), &req)
		replyBack(w, reply, err)
	})))

	mux.Handle(iconst.RoutePathOpen, headerValidator(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req service.OpenRequest
		checkError(proto.Unmarshal(receiveReq(r), &req))
		reply, err := gs.Open(r.Context(), &req)
		replyBack(w, reply, err)
	})))

	mux.Handle(iconst.RoutePathRelease, headerValidator(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req service.ReleaseRequest
		checkError(proto.Unmarshal(receiveReq(r), &req))
		reply, err := gs.Release(r.Context(), &req)
		replyBack(w, reply, err)
	})))

	mux.Handle(iconst.RoutePathReadFile, headerValidator(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req service.ReadFileRequest
		checkError(proto.Unmarshal(receiveReq(r), &req))
		reply, err := gs.ReadFile(r.Context(), &req)
		replyBack(w, reply, err)
	})))

	mux.Handle(iconst.RoutePathStat, headerValidator(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req service.StatRequest
		checkError(proto.Unmarshal(receiveReq(r), &req))
		reply, err := gs.Stat(r.Context(), &req)
		replyBack(w, reply, err)
	})))

	return mux
}
