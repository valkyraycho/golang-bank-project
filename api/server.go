package api

import (
	"fmt"

	db "github.com/valkyraycho/bank_project/db/sqlc"
	"github.com/valkyraycho/bank_project/pb"
	"github.com/valkyraycho/bank_project/token"
	"github.com/valkyraycho/bank_project/utils"
)

type Server struct {
	pb.UnimplementedBankServiceServer
	cfg        utils.Config
	store      db.Store
	tokenMaker token.TokenMaker
}

func NewServer(cfg utils.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(cfg.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create token maker: %w", err)
	}
	return &Server{cfg: cfg, store: store, tokenMaker: tokenMaker}, nil
}
