package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"

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
	// c := pb.NewTextServiceClient(conn)

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
		// uploadResp, err := c.Request(context.Background(), &pb.UploadRequest{Title: "My Upload Title"})
		// if err != nil {
		// 	fmt.Println("Error calling UploadRequest:", err)
		// 	return
		// }

		// Connect to the received port number
		// portStr := fmt.Sprint(uploadResp.GetNumber())
		portStr := "localhost:1234"
		upload(portStr, "test.txt")

	case 2:
		// Call the DownloadRequest RPC method
		portStr := "localhost:1234"
		download(portStr, "test.txt")
	default:
		fmt.Println("Invalid choice. Exiting...")
		return
	}
}

// readFile reads the contents of a file at the specified path and returns it as a byte slice.
func readFile(filePath string) ([]byte, error) {
	// Read the file contents
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return content, nil
}

func upload(address string, filePath string){
    // Open the file for reading
    file, err := os.Open(filePath)
    if err != nil {
        fmt.Println("Error opening file:", err)
        return
    }
    defer file.Close()

    // Connect to the server via TCP
    conn, err := net.Dial("tcp", address)
    if err != nil {
        fmt.Println("Error connecting to server:", err)
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
        // Connect to the server via TCP
    conn, err := net.Dial("tcp", address)
    if err != nil {
        fmt.Println("Error connecting to server:", err)
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

func makeUpload(port string) {
	// Get the file path
	fmt.Println("Provide the path of the file to upload")
	// var filePath string
	// fmt.Scanln(&filePath)
	var ip string = "localhost"
	fmt.Println(ip + ":" + port)
	conn, err := net.Dial("tcp", ip+":"+port)
	if err != nil {
		fmt.Println("Error connecting to port:", err)
		return
	}
	filePath := "test.txt" // Replace with the path to your MP4 file
	fileContent, err := readFile(filePath)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	fmt.Println("File content:")
	fmt.Println(fileContent)

	var currentByte int64 = 0
	fmt.Println("send to client")

	//////////////////////

	// text := "hello world"
	_, err = conn.Write([]byte(fileContent))

	/////////////////////

	fileBuffer := make([]byte, BUFFER_SIZE)

	fmt.Println(currentByte)
	fmt.Println(fileBuffer)

	defer conn.Close()

	fmt.Println("Connected to port:", port)
}
