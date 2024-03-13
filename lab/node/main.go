package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	pb "lab/capitalize"
	"net"
	"strings"
	"sync"
	"time"
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
		time.Sleep(1 * time.Second)
	}
}

func clientCommunication() {
	for {
		// Connect to the client (listening on port 3000)
		listener, err := net.Listen("tcp", "localhost:1234")
		if err != nil {
			fmt.Println("Failed to connect to client:", err)
			time.Sleep(5 * time.Second) // Retry after 5 seconds
			continue
		}
		defer listener.Close()

		conn, err := listener.Accept()

		// Buffer to read incoming data
		buffer := make([]byte, 1024)

		// Read data from client
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("Error reading:", err.Error())
			return
		}

		// Convert received data to string and capitalize it
		receivedText := strings.TrimSpace(string(buffer[:n]))
		capitalizedText := strings.ToUpper(receivedText)

		fmt.Println("capitalizedText:", capitalizedText)

		// Sleep for a while before retrying
		time.Sleep(5 * time.Second)
	}
}
