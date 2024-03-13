package main

import (
	"context"
	"fmt"
	"strconv"

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

	// Ask the user for their choice
	fmt.Println("Press 1 to request upload, or press 2 to request download:")
	var choiceStr string
	fmt.Scanln(&choiceStr)

	// Convert the choice to an integer
	choice, err := strconv.Atoi(choiceStr)
	if err != nil {
		fmt.Println("Invalid choice. Exiting...")
		return
	}

	// Handle user's choice
	switch choice {
	case 1:
		// Call the UploadRequest RPC method
		uploadResp, err := c.Request(context.Background(), &pb.UploadRequest{Title: "My Upload Title"})
		if err != nil {
			fmt.Println("Error calling UploadRequest:", err)
			return
		}
		// Print the received number
		fmt.Println("Received number:", uploadResp.GetNumber())
	case 2:
		// Call the DownloadRequest RPC method
		// Add code here for download request
		fmt.Println("Download request is not implemented yet.")
	default:
		fmt.Println("Invalid choice. Exiting...")
		return
	}
}
