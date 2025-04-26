package grpc

import newsv1 "github.com/codeandlearn1991/news-grpc/api/news/v1"

// Server implements of NewServiceServer.
type Server struct {
	newsv1.UnimplementedNewsServiceServer
}

// NewServer returns an intialized instace of Server.
func NewServer() *Server {
	return &Server{}
}
