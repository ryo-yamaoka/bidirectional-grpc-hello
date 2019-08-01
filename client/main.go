package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"local/grpc/hello"

	"google.golang.org/grpc"
)

func main() {
	conn, _ := grpc.Dial("127.0.0.1:25252", grpc.WithInsecure())
	defer conn.Close()

	c := hello.NewHelloServiceClient(conn)

	resp, _ := c.Greet(context.Background(), &hello.SayHelloMessage{})
	fmt.Println(resp.Msg)

	stream, _ := c.GreetStream(context.Background())

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for {
			msg, err := stream.Recv()
			switch err {
			case io.EOF:
				fmt.Println("Closed")
				return
			case nil:
				fmt.Printf("%s\n", msg.Msg)
			default:
				log.Println(err.Error())
			}
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			if err := stream.Send(&hello.SayHelloMessage{}); err != nil {
				fmt.Println(err.Error())
			}
			time.Sleep(500 * time.Millisecond)
		}
		stream.CloseSend()
	}()

	wg.Wait()
}
