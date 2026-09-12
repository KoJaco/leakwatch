//go:build integration

// gRPC leak generator for E2E tests and manual fixture recording.
package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/KoJaco/leakwatch/test/integration/envutil"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type leakService struct{}

func leakUnaryHandler(_ interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
	var in emptypb.Empty
	if err := dec(&in); err != nil {
		return nil, err
	}
	block := make(chan struct{})
	<-block
	return &emptypb.Empty{}, nil
}

func main() {
	grpcAddr := envutil.Or("LEAKWATCH_GRPC_ADDR", ":50051")
	pprofAddr := envutil.Or("LEAKWATCH_PPROF_ADDR", ":6060")

	go func() {
		fmt.Printf("grpc leak generator: pprof on %s\n", pprofAddr)
		_ = http.ListenAndServe(pprofAddr, nil)
	}()

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		panic(err)
	}

	s := grpc.NewServer()
	s.RegisterService(&grpc.ServiceDesc{
		ServiceName: "leak.LeakService",
		HandlerType: (*any)(nil),
		Methods: []grpc.MethodDesc{{
			MethodName: "Leak",
			Handler:    leakUnaryHandler,
		}},
	}, &leakService{})

	fmt.Printf("grpc leak generator listening on %s\n", grpcAddr)
	go func() {
		if err := s.Serve(lis); err != nil {
			panic(err)
		}
	}()

	time.Sleep(time.Hour)
}
