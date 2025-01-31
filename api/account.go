package api

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	db "github.com/valkyraycho/bank_project/db/sqlc"
	"github.com/valkyraycho/bank_project/pb"
	"github.com/valkyraycho/bank_project/utils"
	"github.com/valkyraycho/bank_project/validator"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) CreateAccount(ctx context.Context, req *pb.CreateAccountRequest) (*pb.CreateAccountResponse, error) {
	payload, err := s.authorizeUser(ctx, utils.SelfAndBanker)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "unauthorized: %s", err)
	}

	violations := validateCreateAccountRequest(req)

	if len(violations) > 0 {
		return nil, invalidArgumentsError(violations)
	}

	if payload.Role != utils.BankerRole && payload.UserID != req.GetOwnerId() {
		return nil, status.Error(codes.PermissionDenied, "no permission to create an account for other users")
	}

	account, err := s.store.CreateAccount(ctx, db.CreateAccountParams{
		OwnerID:  req.GetOwnerId(),
		Currency: req.GetCurrency(),
		Balance:  0,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "user not found: %s", err)
		}
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == db.UniqueViolation {
				return nil, status.Errorf(codes.AlreadyExists, "already exists: %s", err)
			}
		}
		return nil, status.Errorf(codes.Internal, "failed to create account: %s", err)
	}

	return &pb.CreateAccountResponse{Account: convertAccount(account)}, nil
}

func validateCreateAccountRequest(req *pb.CreateAccountRequest) []*errdetails.BadRequest_FieldViolation {
	violations := []*errdetails.BadRequest_FieldViolation{}

	if err := validator.ValidateID(req.OwnerId); err != nil {
		violations = append(violations, fieldViolation("owner_id", err))
	}

	if err := validator.ValidateCurrency(req.Currency); err != nil {
		violations = append(violations, fieldViolation("currency", err))
	}

	return violations
}

func (s *Server) GetAccount(ctx context.Context, req *pb.GetAccountRequest) (*pb.GetAccountResponse, error) {
	payload, err := s.authorizeUser(ctx, utils.SelfAndBanker)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "unauthorized: %s", err)
	}

	violations := validateGetAccountRequest(req)

	if len(violations) > 0 {
		return nil, invalidArgumentsError(violations)
	}

	account, err := s.store.GetAccount(ctx, req.GetId())
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "account not found: %s", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to retrieve account: %s", err)
	}

	if payload.Role != utils.BankerRole && payload.UserID != account.OwnerID {
		return nil, status.Error(codes.PermissionDenied, "no permission to retrieve an account that does not belong to you")
	}

	return &pb.GetAccountResponse{Account: convertAccount(account)}, nil
}

func validateGetAccountRequest(req *pb.GetAccountRequest) []*errdetails.BadRequest_FieldViolation {
	violations := []*errdetails.BadRequest_FieldViolation{}

	if err := validator.ValidateID(req.GetId()); err != nil {
		violations = append(violations, fieldViolation("id", err))
	}
	return violations
}

func (s *Server) GetAccounts(ctx context.Context, req *pb.GetAccountsRequest) (*pb.GetAccountsResponse, error) {
	payload, err := s.authorizeUser(ctx, utils.SelfAndBanker)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "unauthorized: %s", err)
	}

	violations := validateGetAccountsRequest(req)

	if len(violations) > 0 {
		return nil, invalidArgumentsError(violations)
	}

	accounts, err := s.store.ListAccount(ctx, db.ListAccountParams{
		OwnerID: payload.UserID,
		Limit:   req.GetLimit(),
		Offset:  req.GetOffset(),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve account: %s", err)
	}

	if payload.Role != utils.BankerRole && payload.UserID != accounts[0].OwnerID {
		return nil, status.Error(codes.PermissionDenied, "no permission to retrieve accounts that do not belong to you")
	}

	pbAccounts := []*pb.Account{}
	for _, account := range accounts {
		pbAccounts = append(pbAccounts, convertAccount(account))
	}

	return &pb.GetAccountsResponse{Accounts: pbAccounts}, nil
}

func validateGetAccountsRequest(req *pb.GetAccountsRequest) []*errdetails.BadRequest_FieldViolation {
	violations := []*errdetails.BadRequest_FieldViolation{}

	defaultLimit := int32(5)
	defaultOffset := int32(0)

	if req.Limit != nil {
		if err := validator.ValidateLimit(req.GetLimit()); err != nil {
			violations = append(violations, fieldViolation("limit", err))
		}
	} else {
		req.Limit = &defaultLimit
	}

	if req.Offset != nil {
		if err := validator.ValidateOffset(req.GetOffset()); err != nil {
			violations = append(violations, fieldViolation("offset", err))
		}
	} else {
		req.Offset = &defaultOffset
	}

	return violations
}

func convertAccount(account db.Account) *pb.Account {
	return &pb.Account{
		Id:        account.ID,
		OwnerId:   account.OwnerID,
		Balance:   account.Balance,
		Currency:  account.Currency,
		CreatedAt: timestamppb.New(account.CreatedAt),
	}
}
