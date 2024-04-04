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

//works for wifi only
func getAddress()(string){
	ifaces, err := net.Interfaces()
    	if err != nil {
        	fmt.Println(err)
        	return ""
    	}
		var ip string = "" 
    	for _, iface := range ifaces {
        	if iface.Name == "Wi-Fi" {
            	addrs, err := iface.Addrs()
            	if err != nil {
                	fmt.Println(err)
                	return ""
           	 	}

            	for _, addr := range addrs {
                	if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
                    	if ipnet.IP.To4() != nil {
							ip = ipnet.IP.String()
                        	fmt.Println("Current IP address : ", ip)
                    	}
                	}
            	}
        	}
    	}
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

func main() {

	var masterAddress string = "localhost:8080"
	var wg sync.WaitGroup
	wg.Add(2)
	var port string = getAddress()
		var ready chan bool = make(chan bool)
	// Start the server communication thread
	go func() {
		defer wg.Done()
		serverCommunication(port, ready, masterAddress)
	}()

	// Start the client communication thread
	go func() {
		defer wg.Done()
		clientCommunication(port, ready, masterAddress)
	}()

	// Wait for both threads to finish
	wg.Wait()
}



func serverCommunication(port string, ready chan bool, masterAddress string) {
	for {
		// Connect to the gRPC server
		conn, err := grpc.Dial(masterAddress, grpc.WithInsecure())
		if err != nil {
			fmt.Println("Failed to connect to gRPC server:", err)
			time.Sleep(1 * time.Second)
			continue
		}
		// defer conn.Close()

		// Create a gRPC client
		client := pb.NewDataKeeperConnectServiceClient(conn)

		// Perform operations with the server
		// Example: Make an UploadRequest
		// TODO: Modify this to be awake
		// Wait for a value from the ready channel
		r, ok := <-ready
		if !ok {
			fmt.Println("Error: the ready channel was closed")
			return
		}
		if(!r){ // If the server failed to start
			fmt.Println("Error: the server failed to start")
			return
		}
		_, err_1 := client.DataKeeperConnect(context.Background(), &pb.DataKeeperConnectRequest{Port: port})
		if err_1 != nil {
			fmt.Println("Error making UploadRequest:", err)
			time.Sleep(1 * time.Second)
			continue
		}

		// fmt.Println("Received number from server:", uploadResp.GetNumber())

		// Sleep for a while before retrying
		time.Sleep(1000 * time.Second)
	}
}

func clientCommunication(port string, ready chan bool, masterAddress string) {
	listener, err := net.Listen("tcp", "localhost:"+port)
	if err != nil {
		ready <- false
		fmt.Println("Error starting server:", err)
		return
	}
	ready <- true
	defer listener.Close()

	fmt.Println("Node is listening on port localhost:" + port + "...")

	for {
		// Accept client connection
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		// Handle each client connection in a separate goroutine
		go handleClient(conn, port, masterAddress)
	}
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
	fileNameLengthBytes := make([]byte, 10) // assuming filename length is at most 10 bytes
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
	fileSizeBytes := make([]byte, 100) // assuming file size is at most 10 bytes
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

		clientMaster.DataKeeperSuccess(context.Background(), &pb.DataKeeperSuccessRequest{FileName : fileName, DataKeeperNode: port, FilePath :"./" + fileName})

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
