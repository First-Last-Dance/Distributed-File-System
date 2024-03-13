package main

import (
	"context"
	"fmt"
	"net"
	"strings"

	pb "lab/capitalize"

	"google.golang.org/grpc"
)

type textServer struct {
	pb.UnimplementedTextServiceServer
}

func (s *textServer) Capitalize(ctx context.Context, req *pb.TextRequest) (*pb.TextResponse, error) {
	text := req.GetText()
	capitalizedText := strings.ToUpper(text)
	return &pb.TextResponse{CapitalizedText: capitalizedText}, nil
}

func (s *textServer) Request(ctx context.Context, req *pb.UploadRequest) (*pb.UploadResponse, error) {
	return &pb.UploadResponse{Number: 1234}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":7070")
	if err != nil {
		fmt.Println("failed to listen:", err)
		return
	}
	s := grpc.NewServer()
	pb.RegisterTextServiceServer(s, &textServer{})
	fmt.Println("Server started. Listening on port 7070...")
	if err := s.Serve(lis); err != nil {
		fmt.Println("failed to serve:", err)
	}
}
