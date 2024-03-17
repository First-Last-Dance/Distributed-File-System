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
	wg.Add(1)

	// Start the server communication thread
	// go func() {
	// 	defer wg.Done()
	// 	serverCommunication()
	// }()

	// Start the client communication thread
	go func() {
		defer wg.Done()
		clientCommunication()
	}()

	// Wait for both threads to finish
	wg.Wait()
}

func upload(address string, filePath string){
    // Open the file for reading
    file, err := os.Open(filePath)
    if err != nil {
        fmt.Println("Error opening file:", err)
        return
    }
    defer file.Close()

    listener, err := net.Listen("tcp", address)
    if err != nil {
        fmt.Println("Error starting server:", err)
        return
    }
    defer listener.Close()

    fmt.Println("Server is listening on port " + address + "...")

    // Accept client connection
    conn, err := listener.Accept()
    if err != nil {
        fmt.Println("Error accepting connection:", err)
        return
    }
    defer conn.Close()

    // Send the file's contents to the server
    _, err = io.Copy(conn, file)
    if err != nil {
        fmt.Println("Error sending file:", err)
        return
    }

    fmt.Println("File uploaded successfully.")

}
func download(address string, filePath string){
    // Listen for incoming TCP connections on port 8080
	fmt.Println(address)
    listener, err := net.Listen("tcp", address)
    if err != nil {
        fmt.Println("Error starting server:", err)
        return
    }
    defer listener.Close()

    fmt.Println("Server is listening on port " + address + "...")

    // Accept client connection
    conn, err := listener.Accept()
    if err != nil {
        fmt.Println("Error accepting connection:", err)
        return
    }
    defer conn.Close()

    // Create a new file to write the received data
    outFile, err := os.Create(filePath)
    if err != nil {
        fmt.Println("Error creating file:", err)
        return
    }
    defer outFile.Close()

    // Copy the received data to the file
    _, err = io.Copy(outFile, conn)
    if err != nil {
        fmt.Println("Error receiving file:", err)
        return
    }

    fmt.Println("File received and saved as '" + filePath + "'.")

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
	// Connect to the client (listening on port 1234)
	// listener, err := net.Listen("tcp", "localhost:1234")
	// if err != nil {
	// 	fmt.Println("Failed to connect to client:", err)
	// 	time.Sleep(5 * time.Second) // Retry after 5 seconds
	// 	return
	// }
	// defer listener.Close()
	// fmt.Println("Server started. Listening on port 1234...")
	// for {

	// 	conn, err := listener.Accept()
	// 	if err != nil {
	// 		fmt.Println("Error accepting connection:", err.Error())
	// 		return
	// 	}
	// 	// Buffer to read incoming data
	// 	buffer := make([]byte, 1024)

	// 	// Read data from client
	// 	n, err := conn.Read(buffer)
	// 	if err != nil {
	// 		fmt.Println("Error reading:", err.Error())
	// 		return
	// 	}

	// 	// // Convert received data to string and capitalize it
	// 	// receivedText := strings.TrimSpace(string(buffer[:n]))
	// 	// capitalizedText := strings.ToUpper(receivedText)

	// 	fmt.Println("capitalizedText:", buffer[:n])

	// 	// Sleep for a while before retrying
	// 	time.Sleep(5 * time.Second)
	// }
	fmt.Println("Press 1 to request download , or press 2 to request upload:")
	var choiceStr string
	fmt.Scanln(&choiceStr)

	// Convert the choice to an integer
	choice, err := strconv.Atoi(choiceStr)
	if err != nil {
		fmt.Println("Invalid choice. Exiting...")
		return
	}
	switch choice {
	case 1:
		portStr := "localhost:1234"
		download(portStr, "test.txt")

	case 2:
		portStr := "localhost:1234"
		upload(portStr, "test.txt")
	default:
		fmt.Println("Invalid choice. Exiting...")
		return
	}
}
