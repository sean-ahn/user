package server

import (
	"context"
	"net/http"
	"strconv"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/sean-ahn/user/backend/config"
	userv1 "github.com/sean-ahn/user/proto/gen/go/user/v1"
)

func NewHTTPServer(ctx context.Context, cfg config.Config) (*http.Server, error) {
	mux := runtime.NewServeMux(
		runtime.WithMarshalerOption(
			runtime.MIMEWildcard,
			&runtime.JSONPb{
				MarshalOptions: protojson.MarshalOptions{
					UseProtoNames:   true,
					EmitUnpopulated: true,
				},
				UnmarshalOptions: protojson.UnmarshalOptions{
					DiscardUnknown: true,
				},
			},
		),
	)
	options := []grpc.DialOption{
		grpc.WithInsecure(),
	}

	if err := userv1.RegisterUserServiceHandlerFromEndpoint(
		ctx,
		mux,
		":"+strconv.Itoa(cfg.Setting().GRPCServerPort),
		options,
	); err != nil {
		return nil, err
	}

	s := &http.Server{
		Addr:    ":" + strconv.Itoa(cfg.Setting().HTTPServerPort),
		Handler: mux,
	}

	return s, nil
}
