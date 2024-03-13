package main

import (
	"context"
	"fmt"

	pb "lab/capitalize"

	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:7070", grpc.WithInsecure())
	if err != nil {
		fmt.Println("did not connect:", err)
		return
	}
	defer conn.Close()
	c := pb.NewTextServiceClient(conn)

	// Call the UploadRequest RPC method
	uploadResp, err := c.Request(context.Background(), &pb.UploadRequest{Title: "Upload Request"})
	if err != nil {
		fmt.Println("Error calling UploadRequest:", err)
		return
	}

	// Print the received number
	fmt.Println("Received number:", uploadResp.GetNumber())
}
