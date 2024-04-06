package main

import (
	"context"
	"fmt"
	"math/rand"
	"net"

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
	lastUpdate     time.Time
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
	IPAddress := getIPAddress(ctx)
	fileTable = append(fileTable, RowOfFile{request.GetFileName(), IPAddress + ":" + request.GetDataKeeperNode(), request.GetFilePath()})
	if !request.GetIsReplication() {
		replicateFile(request.GetFileName())
	}
	// print fileTable
	fmt.Println("File Table:")
	for _, row := range fileTable {
		fmt.Println(row)
	}
	return &pb.DataKeeperSuccessResponse{}, nil
}

func getIPAddress(ctx context.Context) string {
	pr, ok := peer.FromContext(ctx)
	if !ok {
		// Peer information not available
		return ""
	}
	// Get the client's IP address and port
	clientIP := pr.Addr.(*net.TCPAddr).IP
	return clientIP.String()
}

func (s *server) DataKeeperConnect(ctx context.Context, request *pb.DataKeeperConnectRequest) (*pb.DataKeeperConnectResponse, error) {
	// fmt.Println("DataKeeperConnect called")
	// var node = "Node1"
	// var node = request.GetNode()

	clientIP := getIPAddress(ctx)
	clientPort := request.GetPort()
	var node = clientIP + ":" + clientPort
	for i, row := range nodeTable {
		if row.dataKeeperNode == node {
			nodeTable[i].lastUpdate = time.Now()
			// row.lastUpdate = time.Now()
			// fmt.Println("Node reconnected: ", node)
			// fmt.Println("Node Table:")
			// for _, row := range nodeTable {
			// 	fmt.Println(row)
			// }
			return &pb.DataKeeperConnectResponse{}, nil
		}
	}
	nodeTable = append(nodeTable, RowOfNode{node, time.Now()})
	fmt.Println("Node connected: ", node)
	fmt.Println("Node Table:")
	for _, row := range nodeTable {
		fmt.Println(row)
	}
	return &pb.DataKeeperConnectResponse{}, nil
}

func getAliveNodes() []RowOfNode {
	var aliveNodes []RowOfNode
	currentTime := time.Now()
	for _, node := range nodeTable {
		// fmt.Println("Node Seconds : ", currentTime.Sub(node.lastUpdate).Seconds())
		if currentTime.Sub(node.lastUpdate).Seconds() <= 1.5 {
			aliveNodes = append(aliveNodes, node)
		}
	}
	return aliveNodes
}

func getRandomNode() string {
	aliveNodes := getAliveNodes()
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
			currentTime := time.Now()
			for _, node := range nodeTable {
				// fmt.Println("Time : ", currentTime.Sub(node.lastUpdate).Seconds())
				// fmt.Println("Node : ", node.dataKeeperNode)
				// fmt.Println("Row : ", row.dataKeeperNode)
				if node.dataKeeperNode == row.dataKeeperNode && currentTime.Sub(node.lastUpdate).Seconds() <= 1.5 {
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

func replicateFileFromSourceToDestination(fileName string, sourceNode string, destinationNode string) error {
	conn, err := grpc.Dial(sourceNode, grpc.WithInsecure())
	if err != nil {
		fmt.Println("did not connect to ", sourceNode, " : ", err)
		return err
	}
	defer conn.Close()
	connectToSource := pb.NewDataKeeperReplicateServiceClient(conn)
	_, err = connectToSource.DataKeeperReplicate(context.Background(), &pb.DataKeeperReplicateRequest{FileName: fileName, Node: destinationNode})
	return err
}

func replicateFile(fileName string) {
	nodesContainsFile := getAllNodesContainingFile(fileName)

	if len(nodesContainsFile) == 0 {
		return
	}

	sourceNode := nodesContainsFile[0]
	aliveNodes := getAliveNodes()
	// numOfAliveNodes := len(aliveNodes)
	for {
		if len(nodesContainsFile) >= min(3, len(aliveNodes)) {
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
		err := replicateFileFromSourceToDestination(fileName, sourceNode, destinationNode)
		if err != nil {
			fmt.Println("Error In replication : ", err)
			for i, node := range aliveNodes {
				if node.dataKeeperNode == destinationNode {
					// remove node from alivenodes
					aliveNodes = append(nodeTable[:i], nodeTable[i+1:]...)
					break
				}
			}
		}
		if err == nil {
			fmt.Println("Replication Successfull")
			// nodesContainsFile = append(nodesContainsFile, destinationNode)
			// fileTable = append(fileTable, RowOfFile{fileName, destinationNode, "./" + fileName})
			return
			// fmt.Println("File Table:")
			// for _, row := range fileTable {
			// 	fmt.Println(row)
			// }
		}
	}
}

func replicationAlgorithm() {
	for {
		time.Sleep(20 * time.Second)
		fmt.Println("Replicating files Algo Started")
		distinctFiles := []string{}
		// filesPath := []string{}
		for _, file := range fileTable {
			if !contains(distinctFiles, file.fileName) {
				distinctFiles = append(distinctFiles, file.fileName)
				// filesPath = append(filesPath, file.filePath)
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

	// fileTable = append(fileTable, RowOfFile{"file1", "node1", "/path/to/file1"})
	// fileTable = append(fileTable, RowOfFile{"file1", "node2", "/path/to/file1"})
	// fileTable = append(fileTable, RowOfFile{"file1", "node6", "/path/to/file1"})
	// fileTable = append(fileTable, RowOfFile{"file2", "node2", "/path/to/file2"})
	// fileTable = append(fileTable, RowOfFile{"file3", "node1", "/path/to/file3"})

	// nodeTable = append(nodeTable, RowOfNode{"node1", true})
	// nodeTable = append(nodeTable, RowOfNode{"node2", true})
	// nodeTable = append(nodeTable, RowOfNode{"node6", true})

	lis, err := net.Listen("tcp", "25.23.12.54:8080")
	if err != nil {
		fmt.Println("failed to listen:", err)
		return
	}
	s := grpc.NewServer()
	go replicationAlgorithm()
	pb.RegisterDownloadServiceServer(s, &server{})
	pb.RegisterUploadServiceServer(s, &server{})
	pb.RegisterDataKeeperSuccessServiceServer(s, &server{})
	pb.RegisterDataKeeperConnectServiceServer(s, &server{})
	fmt.Println("Server started")
	if err := s.Serve(lis); err != nil {
		fmt.Println("failed to serve:", err)
	}

}
