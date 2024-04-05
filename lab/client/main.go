package main

import (
	"context"
	"fmt"
	"io"
	pb "lab_1/gRPC"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc"
)

func main() {
	fmt.Println("Press 1 to upload a file or 2 to download a file:")
	var choice int
	fmt.Scanln(&choice)

	var masterAddress string = "25.23.12.54:8080"
	switch choice {
	case 1:
		fmt.Println("Enter file name:")
		var fileName string
		fmt.Scanln(&fileName)
		upload(masterAddress, fileName)
	case 2:
		fmt.Println("Enter file name:")
		var fileName string
		fmt.Scanln(&fileName)
		download(masterAddress, fileName)
	default:
		fmt.Println("Invalid choice")
	}
}

func getAddr(dataKeeperGRPCAddressStr string) string {
	if strings.Count(dataKeeperGRPCAddressStr, ":") > 1 {
    	dataKeeperGRPCAddressStr = "localhost:" + dataKeeperGRPCAddressStr[strings.LastIndex(dataKeeperGRPCAddressStr, ":") + 1:]
		fmt.Println("The address contains more than one colon. assume localhost:port. all ipv6 are assumed to be localhost.")
		fmt.Println("Data keeper gRPC address:", dataKeeperGRPCAddressStr) 
	}
	// Connect to the gRPC server
	conn, err := grpc.Dial(dataKeeperGRPCAddressStr, grpc.WithInsecure())
	if err != nil {
		fmt.Println("Failed to connect to gRPC server:", err)
		return ""
	}
	defer conn.Close()

	// Create a gRPC client
	client := pb.NewDataKeeperOpenConnectionServiceClient(conn)

	
	dataKeeperAddress, err_1 := client.DataKeeperOpenConnection(context.Background(), &pb.DataKeeperOpenConnectionRequest{Empty: ""})
	if err_1 != nil {
		fmt.Println("Error making UploadRequest:", err_1)
		return ""
	}
	dataKeeperAddressStr := dataKeeperAddress.Node
	fmt.Println("Data keeper address:", dataKeeperGRPCAddressStr)
	if strings.Count(dataKeeperAddressStr, ":") > 1 {
    	dataKeeperAddressStr = "localhost:" + dataKeeperAddressStr[strings.LastIndex(dataKeeperAddressStr, ":") + 1:]
		fmt.Println("The address contains more than one colon. assume localhost:port. all ipv6 are assumed to be localhost.")
		fmt.Println("Data keeper gRPC address:", dataKeeperAddressStr) 
	}
	return dataKeeperAddressStr
}

func upload(masterAddress, filePath string) {
	// Connect to the gRPC server
	connMaster, errMaster := grpc.Dial(masterAddress, grpc.WithInsecure())
	if errMaster != nil {
		fmt.Println("Failed to connect to gRPC server:", errMaster)
		time.Sleep(1 * time.Second)
		return
	}
	defer connMaster.Close()

	// Create a gRPC client
	client := pb.NewUploadServiceClient(connMaster)

	// Perform operations with the server
	// Example: Make an UploadRequest
	
	dataKeeperAddress, err_1 := client.Upload(context.Background(), &pb.UploadRequest{Empty: ""})
	if err_1 != nil {
		fmt.Println("Error making UploadRequest:", err_1)
		return
	}
	dataKeeperGRPCAddressStr := dataKeeperAddress.Node
	fmt.Println("Data keeper GRPC address:", dataKeeperGRPCAddressStr)

	dataKeeperAddressStr := getAddr(dataKeeperGRPCAddressStr)
	conn, err := net.Dial("tcp", dataKeeperAddressStr)
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close()

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()



	// Send operation (UPLOAD) => 0
	_, err = conn.Write([]byte{0})
	if err != nil {
		fmt.Println("Error sending operation:", err)
		return
	}


	// Get file info
	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Println("Error getting file info:", err)
		return
	}
	fileSize := fileInfo.Size()
	fileSizeStr := strconv.FormatInt(fileSize, 10)
	
	// Get filename
	fileName := fileInfo.Name()
	fileNameBytes := []byte(fileName)

	// Send length of filename
	fileNameLength := len(fileNameBytes)
	fileNameLengthStr := strconv.Itoa(fileNameLength)
	fileNameLengthStr = fmt.Sprintf("%0100s", fileNameLengthStr)
	_, err = conn.Write([]byte(fileNameLengthStr))
	if err != nil {
		fmt.Println("Error sending filename length:", err)
		return
	}

	// Send filename
	_, err = conn.Write(fileNameBytes)
	if err != nil {
		fmt.Println("Error sending filename:", err)
		return
	}

	// Send file size
	fileSizeStr = fmt.Sprintf("%0100s", fileSizeStr)
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

func download(masterAddress, fileName string) {

	// Connect to the gRPC server
	connMaster, errMaster := grpc.Dial(masterAddress, grpc.WithInsecure())
	if errMaster != nil {
		fmt.Println("Failed to connect to gRPC server:", errMaster)
		time.Sleep(1 * time.Second)
		return
	}
	defer connMaster.Close()

	// Create a gRPC client
	client := pb.NewDownloadServiceClient(connMaster)

	// Perform operations with the server
	// Example: Make an UploadRequest
	
	dataKeeperData, err_1 := client.Download(context.Background(), &pb.DownloadRequest{FileName: fileName})
	if err_1 != nil {
		fmt.Println("Error making UploadRequest:", err_1)
		return
	}
	dataKeeperAddressArr := dataKeeperData.Nodes

	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	// Choose a random dataKeeperAddress
	randomIndex := rand.Intn(len(dataKeeperAddressArr))
	dataKeeperGRPCAddressStr := dataKeeperAddressArr[randomIndex]
	fmt.Println("Data keeper address:", dataKeeperGRPCAddressStr)

	dataKeeperAddressStr := getAddr(dataKeeperGRPCAddressStr)


	if strings.Count(dataKeeperAddressStr, ":") > 1 {
    	dataKeeperAddressStr = "localhost:" + dataKeeperAddressStr[strings.LastIndex(dataKeeperAddressStr, ":") + 1:]
		fmt.Println("The address contains more than one colon. assume localhost:port. all ipv6 are assumed to be localhost.")
		fmt.Println("Data keeper address:", dataKeeperAddressStr) 
	}

	conn, err := net.Dial("tcp", dataKeeperAddressStr)
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close()

	// Send operation (DOWNLOAD) => 1
	_, err = conn.Write([]byte{1})
	if err != nil {
		fmt.Println("Error sending operation:", err)
		return
	}

	fileNameBytes := []byte(fileName)

	// Send length of filename
	fileNameLength := len(fileNameBytes)
	fileNameLengthStr := strconv.Itoa(fileNameLength)
	fileNameLengthStr = fmt.Sprintf("%0100s", fileNameLengthStr)
	_, err = conn.Write([]byte(fileNameLengthStr))
	if err != nil {
		fmt.Println("Error sending filename length:", err)
		return
	}

	// Send filename
	_, err = conn.Write(fileNameBytes)
	if err != nil {
		fmt.Println("Error sending filename:", err)
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
	err = os.WriteFile(fileName , fileContent, 0644)
	if err != nil {
		fmt.Println("Error writing file:", err)
		return
	}

	fmt.Println("File downloaded successfully.")
}
