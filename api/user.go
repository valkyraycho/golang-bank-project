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

func (s *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	violations := validateCreateUserRequest(req)

	if len(violations) > 0 {
		return nil, invalidArgumentsError(violations)
	}

	hashedPassword, err := utils.HashPassword(req.GetPassword())

	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to hash password: %s", err)
	}
	user, err := s.store.CreateUser(ctx, db.CreateUserParams{
		Username:       req.GetUsername(),
		Email:          req.GetEmail(),
		FullName:       req.GetFullName(),
		HashedPassword: hashedPassword,
	})

	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == db.UniqueViolation {
				return nil, status.Errorf(codes.AlreadyExists, "username already exists: %s", err)
			}
		}
		return nil, status.Errorf(codes.Internal, "failed to create user: %s", err)
	}
	return &pb.CreateUserResponse{
		User: &pb.User{
			Username:          user.Username,
			FullName:          user.FullName,
			Email:             user.Email,
			Role:              user.Role,
			PasswordChangedAt: timestamppb.New(user.PasswordChangedAt),
			CreatedAt:         timestamppb.New(user.CreatedAt),
		},
	}, nil
}

func validateCreateUserRequest(req *pb.CreateUserRequest) []*errdetails.BadRequest_FieldViolation {
	violations := []*errdetails.BadRequest_FieldViolation{}

	if err := validator.ValidateUsername(req.Username); err != nil {
		violations = append(violations, fieldViolation("username", err))
	}

	if err := validator.ValidatePassword(req.Password); err != nil {
		violations = append(violations, fieldViolation("password", err))
	}

	if err := validator.ValidateFullName(req.FullName); err != nil {
		violations = append(violations, fieldViolation("full_name", err))
	}

	if err := validator.ValidateEmail(req.Email); err != nil {
		violations = append(violations, fieldViolation("email", err))
	}
	return violations
}

func (s *Server) LoginUser(ctx context.Context, req *pb.LoginUserRequest) (*pb.LoginUserResponse, error) {
	violations := validateLoginUserRequest(req)
	if len(violations) > 0 {
		return nil, invalidArgumentsError(violations)
	}

	user, err := s.store.GetUser(ctx, req.Username)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "user not found: %s", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to find user: %s", err)
	}

	if err := utils.VerifyPassword(req.Password, user.HashedPassword); err != nil {
		return nil, status.Errorf(codes.NotFound, "incorrect password: %s", err)
	}

	accessToken, accessTokenPayload, err := s.tokenMaker.CreateToken(user.ID, user.Role, s.cfg.AccessTokenDuration)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "failed to create access token: %s", err)
	}

	refreshToken, refreshTokenPayload, err := s.tokenMaker.CreateToken(user.ID, user.Role, s.cfg.RefreshTokenDuration)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "failed to create refresh token: %s", err)
	}

	md := s.extractMetadata(ctx)
	session, err := s.store.CreateSession(ctx, db.CreateSessionParams{
		ID:           refreshTokenPayload.ID,
		UserID:       user.ID,
		RefreshToken: refreshToken,
		UserAgent:    md.UserAgent,
		ClientIp:     md.ClientIP,
		IsBlocked:    false,
		ExpiresAt:    refreshTokenPayload.ExpiredAt,
	})
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "failed to create session: %s", err)
	}

	return &pb.LoginUserResponse{
		User: &pb.User{
			Username:          user.Username,
			FullName:          user.FullName,
			Email:             user.Email,
			Role:              user.Role,
			CreatedAt:         timestamppb.New(user.CreatedAt),
			PasswordChangedAt: timestamppb.New(user.PasswordChangedAt),
		},
		SessionId:             session.ID.String(),
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		AccessTokenExpiresAt:  timestamppb.New(accessTokenPayload.ExpiredAt),
		RefreshTokenExpiresAt: timestamppb.New(refreshTokenPayload.ExpiredAt),
	}, nil
}

func validateLoginUserRequest(req *pb.LoginUserRequest) []*errdetails.BadRequest_FieldViolation {
	violations := []*errdetails.BadRequest_FieldViolation{}

	if err := validator.ValidateUsername(req.Username); err != nil {
		violations = append(violations, fieldViolation("username", err))
	}

	if err := validator.ValidatePassword(req.Password); err != nil {
		violations = append(violations, fieldViolation("password", err))
	}
	return violations
}
