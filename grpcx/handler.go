package grpcx

import "google.golang.org/grpc"

// Handler TODO.
type Handler interface{ Register(*grpc.Server) }

// HandlerFunc TODO.
type HandlerFunc func(*grpc.Server)

// Register TODO.
func (hf HandlerFunc) Register(s *grpc.Server) { hf(s) }
