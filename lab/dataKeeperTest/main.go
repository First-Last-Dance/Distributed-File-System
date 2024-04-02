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

func uploadSuccess(c pb.DataKeeperSuccessServiceClient) {

	// Call the RPC method
	fmt.Print("Enter File Name: ")
	var fileName string
	fmt.Scanln(&fileName)

	fmt.Print("Enter Node Name: ")
	var dataKeeperNode string
	fmt.Scanln(&dataKeeperNode)

	fmt.Print("Enter File Path: ")
	var filePath string
	fmt.Scanln(&filePath)

	_, err := c.DataKeeperSuccess(context.Background(), &pb.DataKeeperSuccessRequest{FileName: fileName, DataKeeperNode: dataKeeperNode, FilePath: filePath})
	if err != nil {
		fmt.Println("Error calling Capitalize:", err)
		return
	}
}

func connect(c pb.DataKeeperConnectServiceClient) {

	_, err := c.DataKeeperConnect(context.Background(), &pb.DataKeeperConnectRequest{})
	if err != nil {
		fmt.Println("Error calling Capitalize:", err)
		return
	}
}

func main() {
	conn, err := grpc.Dial("localhost:8080", grpc.WithInsecure())
	if err != nil {
		fmt.Println("did not connect:", err)
		return
	}
	defer conn.Close()
	// c_download := pb.NewDownloadServiceClient(conn)
	// download(c_download)

	// c_upload := pb.NewDataKeeperSuccessServiceClient(conn)
	// uploadSuccess(c_upload)

	c_connect := pb.NewDataKeeperConnectServiceClient(conn)
	connect(c_connect)

}
