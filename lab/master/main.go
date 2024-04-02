package main

import (
	"context"
	"fmt"
	"math/rand"
	"net"

	"strconv"

	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"

	pb "lab_1/gRPC"
)

type RowOfFile struct {
	fileName       string
	dataKeeperNode string
	filePath       string
}

type RowOfNode struct {
	dataKeeperNode string
	isAlive        bool
}

type FileTable []RowOfFile
type NodeTable []RowOfNode

var fileTable FileTable
var nodeTable NodeTable

// Define your service implementation
type server struct {
	pb.UnimplementedDownloadServiceServer
	pb.UnimplementedUploadServiceServer
	pb.DataKeeperSuccessServiceServer
	pb.DataKeeperConnectServiceServer
}

// Implement your gRPC service methods
func (s *server) Download(ctx context.Context, request *pb.DownloadRequest) (*pb.DownloadResponse, error) {
	var fileName = request.GetFileName()
	var nodes = getAllNodesContainingFile(fileName)
	return &pb.DownloadResponse{Nodes: nodes}, nil
}

func (s *server) Upload(ctx context.Context, request *pb.UploadRequest) (*pb.UploadResponse, error) {
	var node = getRandomNode()
	return &pb.UploadResponse{Node: node}, nil
}

func (s *server) DataKeeperSuccess(ctx context.Context, request *pb.DataKeeperSuccessRequest) (*pb.DataKeeperSuccessResponse, error) {
	fileTable = append(fileTable, RowOfFile{request.GetFileName(), request.GetDataKeeperNode(), request.GetFilePath()})
	replicateFile(request.GetFileName())
	// print fileTable
	fmt.Println("File Table:")
	for _, row := range fileTable {
		fmt.Println(row)
	}
	return &pb.DataKeeperSuccessResponse{}, nil
}

func (s *server) DataKeeperConnect(ctx context.Context, request *pb.DataKeeperConnectRequest) (*pb.DataKeeperConnectResponse, error) {
	// fmt.Println("DataKeeperConnect called")
	// var node = "Node1"
	// var node = request.GetNode()
	pr, ok := peer.FromContext(ctx)
	if !ok {
		// Peer information not available
		return nil, fmt.Errorf("failed to get peer information")
	}
	// Get the client's IP address and port
	clientIP := pr.Addr.(*net.TCPAddr).IP
	clientPort := pr.Addr.(*net.TCPAddr).Port
	var node = clientIP.String() + ":" + strconv.Itoa(clientPort)
	nodeTable = append(nodeTable, RowOfNode{node, true})
	fmt.Println("Node connected: ", node)
	// print table
	fmt.Println("Node Table:")
	for _, row := range nodeTable {
		fmt.Println(row)
	}
	return &pb.DataKeeperConnectResponse{}, nil
}

func getRandomNode() string {
	var aliveNodes []RowOfNode
	for _, node := range nodeTable {
		if node.isAlive {
			aliveNodes = append(aliveNodes, node)
		}
	}
	if len(aliveNodes) == 0 {
		return ""
	}
	randomIndex := rand.Intn(len(aliveNodes))
	return aliveNodes[randomIndex].dataKeeperNode
}

func getAllNodesContainingFile(fileName string) []string {
	var nodes []string
	for _, row := range fileTable {
		if row.fileName == fileName {
			for _, node := range nodeTable {
				if node.dataKeeperNode == row.dataKeeperNode && node.isAlive {
					nodes = append(nodes, node.dataKeeperNode)
				}
			}
		}
	}
	return nodes
}

func contains(nodes []string, node string) bool {
	for _, n := range nodes {
		if n == node {
			return true
		}
	}
	return false
}

func replicateFile(fileName string) {
	nodesContainsFile := getAllNodesContainingFile(fileName)
	sourceNode := nodesContainsFile[0]
	for {
		if len(nodesContainsFile) >= 3 {
			break
		}
		var destinationNode string
		for {
			destinationNode = getRandomNode()
			if !contains(nodesContainsFile, destinationNode) {
				break
			}
		}
		fmt.Println("Replicating file", fileName, "from node : ", sourceNode, " to node : ", destinationNode)
		nodesContainsFile = append(nodesContainsFile, destinationNode)
	}
}

func replicationAlgorithm() {
	for {
		time.Sleep(10 * time.Second)
		fmt.Println("Replicating files Algo Started")
		distinctFiles := []string{}
		for _, file := range fileTable {
			if !contains(distinctFiles, file.fileName) {
				distinctFiles = append(distinctFiles, file.fileName)
			}
		}
		// replicate files
		for _, file := range distinctFiles {
			replicateFile(file)
		}
		// fmt.Println("File Table:")
		// for _, row := range fileTable {
		// 	fmt.Println(row)
		// }
	}
}

func main() {

	fileTable = append(fileTable, RowOfFile{"file1", "node1", "/path/to/file1"})
	fileTable = append(fileTable, RowOfFile{"file1", "node2", "/path/to/file1"})
	fileTable = append(fileTable, RowOfFile{"file1", "node6", "/path/to/file1"})
	fileTable = append(fileTable, RowOfFile{"file2", "node2", "/path/to/file2"})
	fileTable = append(fileTable, RowOfFile{"file3", "node1", "/path/to/file3"})

	nodeTable = append(nodeTable, RowOfNode{"node1", true})
	nodeTable = append(nodeTable, RowOfNode{"node2", true})
	nodeTable = append(nodeTable, RowOfNode{"node6", true})

	// go replicationAlgorithm()

	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("failed to listen:", err)
		return
	}
	s := grpc.NewServer()
	pb.RegisterDownloadServiceServer(s, &server{})
	pb.RegisterUploadServiceServer(s, &server{})
	pb.RegisterDataKeeperSuccessServiceServer(s, &server{})
	pb.RegisterDataKeeperConnectServiceServer(s, &server{})
	fmt.Println("Server started")
	if err := s.Serve(lis); err != nil {
		fmt.Println("failed to serve:", err)
	}

}
