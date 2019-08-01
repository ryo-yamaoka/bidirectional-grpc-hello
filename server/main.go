package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"local/grpc/hello"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"google.golang.org/grpc"
)

type helloService struct{}

func (s helloService) Greet(ctx context.Context, _ *hello.SayHelloMessage) (*hello.SayHelloResponse, error) {
	return &hello.SayHelloResponse{
		Msg: "Hello, gRPC",
	}, nil
}

func (s helloService) GreetStream(stream hello.HelloService_GreetStreamServer) error {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		max := 10
		for i := 1; i <= max; i++ {
			stream.Send(&hello.SayHelloResponse{
				Msg: fmt.Sprintf("Hello, gRPC %d/%d", i, max),
			})
			time.Sleep(1 * time.Second)
		}
	}()

	go func() {
		defer wg.Done()
		var receivedCount int
		for {
			_, err := stream.Recv()
			receivedCount++
			switch err {
			case io.EOF:
				fmt.Println("Closed")
				return
			case nil:
				fmt.Printf("received: %d\n", receivedCount)
			default:
				fmt.Println(err.Error())
				return
			}
		}
	}()

	wg.Wait()
	return nil
}

func main() {
	s := grpc.NewServer()
	hello.RegisterHelloServiceServer(s, &helloService{})
	port, _ := net.Listen("tcp", ":25252")

	flag.Parse()
	args := flag.Args()
	switch {
	case len(args) == 0:
		s.Serve(port)
	case args[0] == "web":
		gRPCWebMode(s, port)
	default:
		fmt.Println("Unknown mode")
	}
}

func gRPCWebMode(s *grpc.Server, port net.Listener) {
	wrappedServer := grpcweb.WrapServer(
		s,
		grpcweb.WithOriginFunc(func(_ string) bool { return true }),
	)

	hs := http.Server{
		Handler: http.HandlerFunc(wrappedServer.ServeHTTP),
	}

	hs.Serve(port)
}
