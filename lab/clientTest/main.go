package main

import (
	"context"
	"fmt"

	pb "lab_1/gRPC"

	"google.golang.org/grpc"
)

func download(c pb.DownloadServiceClient) {

	// Read input from user
	fmt.Print("Enter File Name: ")
	var fileName string
	fmt.Scanln(&fileName)

	// Call the RPC method
	resp, err := c.Download(context.Background(), &pb.DownloadRequest{FileName: fileName})
	if err != nil {
		fmt.Println("Error calling Capitalize:", err)
		return
	}

	// Print the result
	fmt.Println("Nodes:", resp.GetNodes())
}

func upload(c pb.UploadServiceClient) {

	// Call the RPC method
	resp, err := c.Upload(context.Background(), &pb.UploadRequest{})
	if err != nil {
		fmt.Println("Error calling Capitalize:", err)
		return
	}

	// Print the result
	fmt.Println("Node:", resp.GetNode())
}

func main() {
	conn, err := grpc.Dial("localhost:8080", grpc.WithInsecure())
	if err != nil {
		fmt.Println("did not connect:", err)
		return
	}
	defer conn.Close()
	c_download := pb.NewDownloadServiceClient(conn)
	download(c_download)

	// c_upload := pb.NewUploadServiceClient(conn)
	// upload(c_upload)

}
