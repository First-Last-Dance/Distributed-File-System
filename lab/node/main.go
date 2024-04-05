package main

import (
	"context"
	"fmt"
	"io"
	pb "lab_1/gRPC"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"
)

func getPort() string {
	// Get a free port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		fmt.Println("Error while listening for a free port:", err)
		return ""
	}
	var port int = listener.Addr().(*net.TCPAddr).Port
	fmt.Println("Free port:", port)
	listener.Close()

	return strconv.Itoa(port)
}


type server struct {
	pb.UnimplementedDataKeeperReplicateServiceServer
	pb.UnimplementedDataKeeperOpenConnectionServiceServer
}

func (s *server) DataKeeperReplicate(ctx context.Context, request *pb.DataKeeperReplicateRequest) (*pb.DataKeeperReplicateResponse, error) {
	var fileName = request.GetFileName()
	var node = request.GetNode()
	go replicateFile(fileName, node)
	return &pb.DataKeeperReplicateResponse{Empty_Response: ""}, nil
}

func (s *server) DataKeeperOpenConnection(ctx context.Context, request *pb.DataKeeperOpenConnectionRequest) (*pb.DataKeeperOpenConnectionResponse, error) {
	var port = getPort()
	go clientCommunication(dataKeeperIP, port, masterAddress)
	return &pb.DataKeeperOpenConnectionResponse{Node: dataKeeperIP+":"+port}, nil
}

func replicateFile(filePath string, node string) {
	dataKeeperAddressStr := node
	fmt.Println("Data keeper address:", dataKeeperAddressStr)

	if strings.Count(dataKeeperAddressStr, ":") > 1 {
    	dataKeeperAddressStr = "localhost:" + dataKeeperAddressStr[strings.LastIndex(dataKeeperAddressStr, ":") + 1:]
		fmt.Println("The address contains more than one colon. assume localhost:port. all ipv6 are assumed to be localhost.")
		fmt.Println("Data keeper address:", dataKeeperAddressStr) 
	}

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	conn, err := net.Dial("tcp", dataKeeperAddressStr)
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close()

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

var masterAddress string = "25.23.12.54:8080"
var dataKeeperIP string = "25.49.63.207"

func main() {

	var wg sync.WaitGroup
	wg.Add(1)
	var port string = getPort()

	listener, err := net.Listen("tcp", dataKeeperIP + ":" +port)
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()
	s := grpc.NewServer()
	pb.RegisterDataKeeperReplicateServiceServer(s, &server{})
	pb.RegisterDataKeeperOpenConnectionServiceServer(s, &server{})

	if err := s.Serve(listener); err != nil {
		fmt.Println("failed to serve:", err)
	}
	fmt.Println("Node is listening on port 25.49.63.207:" + port + "...") 
	// Start the server communication thread
	go func() {
		defer wg.Done()
		serverCommunication(port, masterAddress)
	}()

	// Wait for thread to finish
	wg.Wait()
}

func serverCommunication(port string, masterAddress string) {
	var conn *grpc.ClientConn
	defer conn.Close()
	var client pb.DataKeeperConnectServiceClient
	for {
		// Connect to the gRPC server
		conn, err := grpc.Dial(masterAddress, grpc.WithInsecure())
		if err != nil {
			fmt.Println("Failed to connect to gRPC server:", err)
			time.Sleep(1 * time.Second)
			continue
		}
		// Create a gRPC client
		client = pb.NewDataKeeperConnectServiceClient(conn)
		_, err_1 := client.DataKeeperConnect(context.Background(), &pb.DataKeeperConnectRequest{Port: port})
		if err_1 != nil {
			fmt.Println("Error connecting to  master:", err_1)
			time.Sleep(time.Second)
			continue
		}else{
			fmt.Println("Connected to master")
			break
		}
	}

	for {
		time.Sleep(time.Second)
		_, err_1 := client.DataKeeperConnect(context.Background(), &pb.DataKeeperConnectRequest{Port: port})
		if err_1 != nil {
			fmt.Println("Error reconnecting to master:", err_1)
			continue
		}else{
			continue
		}

	}
}

func clientCommunication(ip string, port string, masterAddress string) {
	listener, err := net.Listen("tcp", ip+":"+port)
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()
	fmt.Println("Node is listening on port " + dataKeeperIP + ":" + port + "...") 
	// Accept client connection
	conn, err := listener.Accept()
	if err != nil {
		fmt.Println("Error accepting connection:", err)
	}

	// Handle each client connection in a separate goroutine
	go handleClient(conn, port, masterAddress)
}

func handleClient(conn net.Conn, port string, masterAddress string) {
	defer conn.Close()

	// Read the first byte to determine operation (UPLOAD or DOWNLOAD)
	opBuffer := make([]byte, 1)
	_, err := conn.Read(opBuffer)
	if err != nil {
		fmt.Println("Error reading operation:", err)
		return
	}

	op := opBuffer[0]

	switch op {
	case 0:
		download(conn, port, masterAddress)
	case 1:
		upload(conn)
	default:
		fmt.Println("Unknown operation:", op)
	}
}

func download(conn net.Conn, port string, masterAddress string) {
	// Receive length of filename
	fileNameLengthBytes := make([]byte, 100) // assuming filename length is at most 100 bytes
	_, err := io.ReadFull(conn, fileNameLengthBytes)
	if err != nil {
		fmt.Println("Error receiving filename length:", err)
		return
	}
	fileNameLengthStr := strings.TrimSpace(string(fileNameLengthBytes))
	fileNameLength, err := strconv.Atoi(fileNameLengthStr)
	if err != nil {
		fmt.Println("Error converting filename length to integer:", err)
		return
	}
	// Receive filename
	fileNameBytes := make([]byte, fileNameLength)
	_, err = io.ReadFull(conn, fileNameBytes)
	if err != nil {
		fmt.Println("Error receiving filename:", err)
		return
	}
	fileName := string(fileNameBytes)

	// Receive file size
	fileSizeBytes := make([]byte, 100) // assuming file size is at most 100 bytes
	_, err = io.ReadFull(conn, fileSizeBytes)
	if err != nil {
		fmt.Println("Error receiving file size:", err)
		return
	}
	fileSize, _ := strconv.Atoi(string(fileSizeBytes))

	// Receive file content
	fileContent := make([]byte, fileSize)
	_, err = io.ReadFull(conn, fileContent)
	if err != nil {
		fmt.Println("Error receiving file content:", err)
		return
	}

	// Write received content to file
	err = os.WriteFile(fileName, fileContent, 0644)
	if err != nil {
		fmt.Println("Error writing file:", err)
		return
	}

	connMaster, err := grpc.Dial(masterAddress, grpc.WithInsecure())
	if err != nil {
		fmt.Println("Failed to connect to gRPC server:", err)
		time.Sleep(1 * time.Second)
		return
	}
	defer connMaster.Close()

	// Create a gRPC client
	clientMaster := pb.NewDataKeeperSuccessServiceClient(connMaster)

	clientMaster.DataKeeperSuccess(context.Background(), &pb.DataKeeperSuccessRequest{FileName: fileName, DataKeeperNode: port, FilePath: "./" + fileName, IsReplication: false})

	fmt.Println("File uploaded successfully.")

}

func upload(conn net.Conn) {
	// Receive length of filePath
	filePathLengthBytes := make([]byte, 100) // assuming filePath length is at most 100 bytes
	_, err := io.ReadFull(conn, filePathLengthBytes)
	if err != nil {
		fmt.Println("Error receiving filePath length:", err)
		return
	}
	filePathLengthStr := strings.TrimSpace(string(filePathLengthBytes))
	filePathLength, err := strconv.Atoi(filePathLengthStr)
	if err != nil {
		fmt.Println("Error converting filePath length to integer:", err)
		return
	}
	// Receive filePath
	filePathBytes := make([]byte, filePathLength)
	_, err = io.ReadFull(conn, filePathBytes)
	if err != nil {
		fmt.Println("Error receiving filePath:", err)
		return
	}
	filePath := string(filePathBytes)
	// Open the file for reading
	file, err := os.Open(filePath)
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
