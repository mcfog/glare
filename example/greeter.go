package main

import (
	"log"
	"time"

	"github.com/mcfog/glare/pkg/glare"
	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
)

const (
	address = "localhost:50051"
)

func main() {
	// connect to your grpc endpoint
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	// register clients in string map
	clientMap := make(map[string]interface{})
	clientMap["greeter"] = pb.NewGreeterClient(conn)

	// instantiat ReflectServer, pay attention to timeout
	svr := glare.NewReflectServer(clientMap, time.Second * 3)

	// Listen!!
	svr.Run("tcp", ":6380")
}
