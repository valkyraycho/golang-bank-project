package api

import (
	"context"

	"github.com/jackc/pgx/v5"
	db "github.com/valkyraycho/bank_project/db/sqlc"
	"github.com/valkyraycho/bank_project/pb"
	"github.com/valkyraycho/bank_project/utils"
	"github.com/valkyraycho/bank_project/validator"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) CreateTransfer(ctx context.Context, req *pb.CreateTransferRequest) (*pb.CreateTransferResponse, error) {
	payload, err := s.authorizeUser(ctx, []string{utils.CustomerRole})
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "unauthorized: %s", err)
	}

	violations := validateCreateTransferRequest(req)
	if len(violations) > 0 {
		return nil, invalidArgumentsError(violations)
	}

	if req.FromAccountId == req.ToAccountId {
		return nil, status.Errorf(codes.InvalidArgument, "cannot transfer to the same account")
	}

	fromAccount, err := s.store.GetAccount(ctx, req.FromAccountId)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "account not found: %s", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to retrieve account: %s", err)
	}

	if payload.UserID != fromAccount.OwnerID {
		return nil, status.Error(codes.PermissionDenied, "no permission to transfer from this account")
	}

	toAccount, err := s.store.GetAccount(ctx, req.ToAccountId)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "account not found: %s", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to retrieve account: %s", err)
	}

	// Check currency match for both accounts
	if fromAccount.Currency != req.Currency || toAccount.Currency != req.Currency {
		return nil, status.Errorf(codes.FailedPrecondition, "account currency must match transfer currency %s", req.Currency)
	}

	res, err := s.store.TransferTx(ctx, db.TransferTxParams{
		FromAccountID: req.FromAccountId,
		ToAccountID:   req.ToAccountId,
		Amount:        req.Amount,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create account: %s", err)
	}

	return &pb.CreateTransferResponse{
		Transfer: &pb.Transfer{
			FromAccountId: res.Transfer.FromAccountID,
			ToAccountId:   res.Transfer.ToAccountID,
			Amount:        res.Transfer.Amount,
			CreatedAt:     timestamppb.New(res.Transfer.CreatedAt),
		},
		FromAccount: &pb.Account{
			Id:        res.FromAccount.ID,
			OwnerId:   res.FromAccount.OwnerID,
			Balance:   res.FromAccount.Balance,
			Currency:  res.FromAccount.Currency,
			CreatedAt: timestamppb.New(res.FromAccount.CreatedAt),
		},
		ToAccount: &pb.Account{
			Id:        res.ToAccount.ID,
			OwnerId:   res.ToAccount.OwnerID,
			Balance:   res.ToAccount.Balance,
			Currency:  res.ToAccount.Currency,
			CreatedAt: timestamppb.New(res.ToAccount.CreatedAt),
		},
		FromEntry: &pb.Entry{
			AccountId: res.FromEntry.AccountID,
			Amount:    res.FromEntry.Amount,
			CreatedAt: timestamppb.New(res.FromEntry.CreatedAt),
		},
		ToEntry: &pb.Entry{
			AccountId: res.ToEntry.AccountID,
			Amount:    res.ToEntry.Amount,
			CreatedAt: timestamppb.New(res.ToEntry.CreatedAt),
		},
	}, nil
}

func validateCreateTransferRequest(req *pb.CreateTransferRequest) []*errdetails.BadRequest_FieldViolation {
	violations := []*errdetails.BadRequest_FieldViolation{}

	if err := validator.ValidateID(req.GetFromAccountId()); err != nil {
		violations = append(violations, fieldViolation("from_account_id", err))
	}

	if err := validator.ValidateID(req.GetToAccountId()); err != nil {
		violations = append(violations, fieldViolation("to_account_id", err))
	}

	if err := validator.ValidateAmount(req.Amount); err != nil {
		violations = append(violations, fieldViolation("amount", err))
	}

	return violations
}
