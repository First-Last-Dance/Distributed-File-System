package main

import (
	"context"
	"fmt"
	pb "lab/capitalize"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"google.golang.org/grpc"
)

const BUFFER_SIZE = 1024

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

		// Connect to the received port number
		portStr := fmt.Sprint(uploadResp.GetNumber())
		makeUpload(portStr)

	case 2:
		// Call the DownloadRequest RPC method
		fmt.Println("Download request is not implemented yet.")
	default:
		fmt.Println("Invalid choice. Exiting...")
		return
	}
}

func makeUpload(port string) {
	// Get the file path
	fmt.Println("Provide the path of the file to upload")
	var filePath string
	fmt.Scanln(&filePath)
	var ip string = "localhost"
	fmt.Println(ip + ":" + port)
	conn, err := net.Dial("tcp", ip+":"+port)
	if err != nil {
		fmt.Println("Error connecting to port:", err)
		return
	}

	var currentByte int64 = 0
	fmt.Println("send to client")

	//////////////////////

	text := "hello world"
	_, err = conn.Write([]byte(strings.TrimSpace(text)))

	/////////////////////

	fileBuffer := make([]byte, BUFFER_SIZE)

	var error error

	//file to read
	file, error := os.Open(strings.TrimSpace(filePath)) // For read access.
	if error != nil {
		conn.Write([]byte("-1"))
		log.Fatal(error)
	}

	fmt.Println(currentByte)
	fmt.Println(fileBuffer)
	fmt.Println(file)

	defer conn.Close()

	fmt.Println("Connected to port:", port)
}
