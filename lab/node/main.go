package main

import (
	"context"
	"fmt"
	"io"
	pb "lab/capitalize"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"google.golang.org/grpc"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(2)

	// Start the server communication thread
	go func() {
		defer wg.Done()
		serverCommunication()
	}()

	// Start the client communication thread
	go func() {
		defer wg.Done()
		clientCommunication()
	}()

	// Wait for both threads to finish
	wg.Wait()
}

//////////////////////////////////////////////////////////////////

func serverCommunication() {
	for {
		// Connect to the gRPC server
		conn, err := grpc.Dial("localhost:7070", grpc.WithInsecure())
		if err != nil {
			fmt.Println("Failed to connect to gRPC server:", err)
			time.Sleep(1 * time.Second)
			continue
		}
		// defer conn.Close()

		// Create a gRPC client
		client := pb.NewTextServiceClient(conn)

		// Perform operations with the server
		// Example: Make an UploadRequest
		// TODO: Modify this to be awake
		uploadResp, err := client.Request(context.Background(), &pb.UploadRequest{Title: "My Upload Title"})
		if err != nil {
			fmt.Println("Error making UploadRequest:", err)
			time.Sleep(1 * time.Second)
			continue
		}

		fmt.Println("Received number from server:", uploadResp.GetNumber())

		// Sleep for a while before retrying
		time.Sleep(1000 * time.Second)
	}
}

// ////////////////////////////////////////////////////////////////
func clientCommunication() {
	address := "localhost:1234"
	listener, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Node is listening on port " + address + "...")

	for {
		// Accept client connection
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		// Handle each client connection in a separate goroutine
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()

	// Read the first byte to determine operation (UPLOAD or DOWNLOAD)
	opBuffer := make([]byte, 10)
	n, err := conn.Read(opBuffer)
	if err != nil {
		fmt.Println("Error reading operation:", err)
		return
	}

	op := string(opBuffer[:n])

	switch op {
	case "UPLOAD":
		download(conn)
	case "DOWNLOAD":
		upload(conn)
	default:
		fmt.Println("Unknown operation:", op)
	}
}

func download(conn net.Conn) {
	// Receive file size
	sizeBuffer := make([]byte, 64)
	n, err := conn.Read(sizeBuffer)
	if err != nil {
		fmt.Println("Error reading file size:", err)
		return
	}

	sizeStr := string(sizeBuffer[:n])
	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		fmt.Println("Error parsing file size:", err)
		return
	}

	// Receive file content
	fileContent := make([]byte, size)
	_, err = io.ReadFull(conn, fileContent)
	if err != nil {
		fmt.Println("Error receiving file content:", err)
		return
	}

	// Write received content to file
	err = os.WriteFile("video.mkv", fileContent, 0644)
	if err != nil {
		fmt.Println("Error writing file:", err)
		return
	}

	fmt.Println("File uploaded successfully.")
}

func upload(conn net.Conn) {
	// Open the file for reading
	file, err := os.Open("test.txt")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// Get file size
	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Println("Error getting file info:", err)
		return
	}
	fileSize := fileInfo.Size()

	// Send file size to client
	fileSizeStr := strconv.FormatInt(fileSize, 10)
	_, err = conn.Write([]byte(fileSizeStr))
	if err != nil {
		fmt.Println("Error sending file size:", err)
		return
	}

	// Send file content to client
	_, err = io.Copy(conn, file)
	if err != nil {
		fmt.Println("Error sending file content:", err)
		return
	}

	fmt.Println("File sent successfully.")
}
