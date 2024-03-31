package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
)

func main() {
	fmt.Println("Press 1 to upload a file or 2 to download a file:")
	var choice int
	fmt.Scanln(&choice)

	switch choice {
	case 1:
		fmt.Println("Enter file name:")
		var fileName string
		fmt.Scanln(&fileName)
		upload("localhost:1234", fileName)
	case 2:
		download("localhost:1234", "downloaded_file.txt")
	default:
		fmt.Println("Invalid choice")
	}
}

func upload(address, filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close()

	// Send operation (UPLOAD)
	_, err = conn.Write([]byte("UPLOAD"))
	if err != nil {
		fmt.Println("Error sending operation:", err)
		return
	}

	// Send file size
	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Println("Error getting file info:", err)
		return
	}
	fileSize := fileInfo.Size()
	fileSizeStr := strconv.FormatInt(fileSize, 10)
	_, err = conn.Write([]byte(fileSizeStr))
	if err != nil {
		fmt.Println("Error sending file size:", err)
		return
	}

	// Send file content
	_, err = io.Copy(conn, file)
	if err != nil {
		fmt.Println("Error sending file content:", err)
		return
	}

	fmt.Println("File uploaded successfully.")
}

func download(address, filePath string) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close()

	// Send operation (DOWNLOAD)
	_, err = conn.Write([]byte("DOWNLOAD"))
	if err != nil {
		fmt.Println("Error sending operation:", err)
		return
	}

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
	err = os.WriteFile(filePath, fileContent, 0644)
	if err != nil {
		fmt.Println("Error writing file:", err)
		return
	}

	fmt.Println("File downloaded successfully.")
}
