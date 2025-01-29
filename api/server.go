package api

import (
	db "github.com/valkyraycho/bank_project/db/sqlc"
	"github.com/valkyraycho/bank_project/pb"
)

type Server struct {
	pb.UnimplementedBankServiceServer
	store db.Store
}

func NewServer(store db.Store) *Server {
	return &Server{store: store}
}
