package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"local/grpc/hello"

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
	port, _ := net.Listen("tcp", ":25252")
	s := grpc.NewServer()
	hello.RegisterHelloServiceServer(s, &helloService{})
	s.Serve(port)
}
