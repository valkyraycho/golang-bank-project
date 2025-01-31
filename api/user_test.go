package api

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	mockdb "github.com/valkyraycho/bank_project/db/mock"
	db "github.com/valkyraycho/bank_project/db/sqlc"
	"github.com/valkyraycho/bank_project/pb"
	"github.com/valkyraycho/bank_project/token"
	"github.com/valkyraycho/bank_project/utils"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type eqCreateUserParamsMatcher struct {
	args     db.CreateUserParams
	password string
}

func (matcher eqCreateUserParamsMatcher) Matches(x any) bool {
	args, ok := x.(db.CreateUserParams)
	if !ok {
		return false
	}
	if err := utils.VerifyPassword(matcher.password, args.HashedPassword); err != nil {
		return false
	}
	matcher.args.HashedPassword = args.HashedPassword

	if !reflect.DeepEqual(matcher.args, args) {
		return false
	}
	return true
}

func (matcher eqCreateUserParamsMatcher) String() string {
	return fmt.Sprintf("matches arg %v and password %v", matcher.args, matcher.password)
}

func eqCreateUserParams(args db.CreateUserParams, password string) gomock.Matcher {
	return eqCreateUserParamsMatcher{args: args, password: password}
}

func TestCreateUser(t *testing.T) {
	user, password := randomUser(t)

	testCases := []struct {
		name          string
		req           *pb.CreateUserRequest
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, res *pb.CreateUserResponse, err error)
	}{
		{
			name: "OK",
			req: &pb.CreateUserRequest{
				Username: user.Username,
				Password: password,
				FullName: user.FullName,
				Email:    user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), eqCreateUserParams(db.CreateUserParams{
						Username: user.Username,
						FullName: user.FullName,
						Email:    user.Email,
					}, password)).
					Times(1).
					Return(user, nil)
			},
			checkResponse: func(t *testing.T, res *pb.CreateUserResponse, err error) {
				require.NoError(t, err)
				require.NotEmpty(t, res)
				createdUser := res.GetUser()
				require.NotEmpty(t, createdUser)
				require.Equal(t, user.Username, createdUser.Username)
				require.Equal(t, user.Email, createdUser.Email)
				require.Equal(t, user.FullName, createdUser.FullName)
			},
		},
		{
			name: "InternalError",
			req: &pb.CreateUserRequest{
				Username: user.Username,
				Password: password,
				FullName: user.FullName,
				Email:    user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, res *pb.CreateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Internal, st.Code())
			},
		},
		{
			name: "DuplicateUsername",
			req: &pb.CreateUserRequest{
				Username: user.Username,
				Password: password,
				FullName: user.FullName,
				Email:    user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, db.ErrUniqueViolation)
			},
			checkResponse: func(t *testing.T, res *pb.CreateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.AlreadyExists, st.Code())
			},
		},
		{
			name: "InvalidEmail",
			req: &pb.CreateUserRequest{
				Username: user.Username,
				Password: password,
				FullName: user.FullName,
				Email:    "invalid",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, res *pb.CreateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.InvalidArgument, st.Code())
			},
		},
		{
			name: "InvalidUsername",
			req: &pb.CreateUserRequest{
				Username: "*%(#$A$#(@))",
				Password: password,
				FullName: user.FullName,
				Email:    user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, res *pb.CreateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.InvalidArgument, st.Code())
			},
		},
		{
			name: "InvalidFullName",
			req: &pb.CreateUserRequest{
				Username: user.Username,
				Password: password,
				FullName: "*%(#$A$#(@))",
				Email:    user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, res *pb.CreateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.InvalidArgument, st.Code())
			},
		},
		{
			name: "InvalidPassword",
			req: &pb.CreateUserRequest{
				Username: user.Username,
				Password: "abc",
				FullName: user.FullName,
				Email:    user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, res *pb.CreateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.InvalidArgument, st.Code())
			},
		},
	}

	for _, testCase := range testCases {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		store := mockdb.NewMockStore(ctrl)

		testCase.buildStubs(store)

		server := NewTestServer(t, store)
		res, err := server.CreateUser(context.Background(), testCase.req)
		testCase.checkResponse(t, res, err)
	}
}

func TestLoginUser(t *testing.T) {
	user, password := randomUser(t)

	testCases := []struct {
		name          string
		req           *pb.LoginUserRequest
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, res *pb.LoginUserResponse, err error)
	}{
		{
			name: "OK",
			req: &pb.LoginUserRequest{
				Username: user.Username,
				Password: password,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.Username)).
					Times(1).
					Return(user, nil)

				store.EXPECT().
					CreateSession(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Session{
						ID:           utils.RandomUUID(),
						UserID:       user.ID,
						RefreshToken: utils.RandomString(32),
						UserAgent:    "",
						ClientIp:     "",
						IsBlocked:    false,
						ExpiresAt:    time.Now().Add(time.Hour * 24),
					}, nil)
			},
			checkResponse: func(t *testing.T, res *pb.LoginUserResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, res)

				gotUser := res.GetUser()
				require.Equal(t, user.Username, gotUser.Username)
				require.Equal(t, user.Email, gotUser.Email)
				require.Equal(t, user.FullName, gotUser.FullName)

				require.NotEmpty(t, res.GetSessionId())
				require.NotEmpty(t, res.GetAccessToken())
				require.NotEmpty(t, res.GetRefreshToken())

				require.True(t, res.GetAccessToken() != res.GetRefreshToken())
			},
		},
		{
			name: "InternalError",
			req: &pb.LoginUserRequest{
				Username: user.Username,
				Password: password,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, sql.ErrConnDone)

				store.EXPECT().
					CreateSession(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, res *pb.LoginUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Internal, st.Code())
			},
		},
		{
			name: "UserNotFound",
			req: &pb.LoginUserRequest{
				Username: "notfound",
				Password: password,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, pgx.ErrNoRows)

				store.EXPECT().
					CreateSession(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, res *pb.LoginUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.NotFound, st.Code())
			},
		},
		{
			name: "IncorrectPassword",
			req: &pb.LoginUserRequest{
				Username: user.Username,
				Password: "incorrect",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.Username)).
					Times(1).
					Return(user, nil)

				store.EXPECT().
					CreateSession(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, res *pb.LoginUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.NotFound, st.Code())
			},
		},
		{
			name: "InvalidUsername",
			req: &pb.LoginUserRequest{
				Username: "*%(#$A$#(@))",
				Password: password,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(0)

				store.EXPECT().
					CreateSession(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, res *pb.LoginUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.InvalidArgument, st.Code())
			},
		},
		{
			name: "InvalidPassword",
			req: &pb.LoginUserRequest{
				Username: user.Username,
				Password: "abc",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(0)

				store.EXPECT().
					CreateSession(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, res *pb.LoginUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.InvalidArgument, st.Code())
			},
		},
	}

	for _, testCase := range testCases {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		store := mockdb.NewMockStore(ctrl)

		testCase.buildStubs(store)

		server := NewTestServer(t, store)
		res, err := server.LoginUser(context.Background(), testCase.req)
		testCase.checkResponse(t, res, err)
	}
}

func TestUpdateUser(t *testing.T) {
	user, _ := randomUser(t)

	newName := utils.RandomName()
	newEmail := utils.RandomEmail()

	invalidUsername := "*%(#$A$#(@))"
	invalidFullName := "*%(#$A$#(@))"
	invalidEmail := "invalid_email"
	InvalidPassword := "abc"

	testCases := []struct {
		name          string
		req           *pb.UpdateUserRequest
		buildStubs    func(store *mockdb.MockStore)
		buildContext  func(t *testing.T, tokenMaker token.TokenMaker) context.Context
		checkResponse func(t *testing.T, res *pb.UpdateUserResponse, err error)
	}{
		{
			name: "OK",
			req: &pb.UpdateUserRequest{
				Id:       user.ID,
				Username: &newName,
				Email:    &newEmail,
				FullName: &newName,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Eq(db.UpdateUserParams{
						ID: user.ID,
						Username: pgtype.Text{
							String: newName,
							Valid:  true,
						},
						Email: pgtype.Text{
							String: newEmail,
							Valid:  true,
						},
						FullName: pgtype.Text{
							String: newName,
							Valid:  true,
						},
					})).
					Times(1).
					Return(db.User{
						ID:                user.ID,
						Username:          newName,
						HashedPassword:    user.HashedPassword,
						FullName:          newName,
						Email:             newEmail,
						PasswordChangedAt: user.PasswordChangedAt,
						CreatedAt:         user.CreatedAt,
					}, nil)
			},
			buildContext: func(t *testing.T, tokenMaker token.TokenMaker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user.ID, user.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
				require.NoError(t, err)
				require.NotEmpty(t, res)
				updatedUser := res.GetUser()
				require.NotEmpty(t, updatedUser)
				require.Equal(t, newName, updatedUser.Username)
				require.Equal(t, newEmail, updatedUser.Email)
				require.Equal(t, newName, updatedUser.FullName)
			},
		},
		{
			name: "InternalError",
			req: &pb.UpdateUserRequest{
				Id:       user.ID,
				Username: &newName,
				Email:    &newEmail,
				FullName: &newName,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, sql.ErrConnDone)
			},
			buildContext: func(t *testing.T, tokenMaker token.TokenMaker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user.ID, user.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Internal, st.Code())
			},
		},
		{
			name: "UserNotFound",
			req: &pb.UpdateUserRequest{
				Id:       user.ID,
				Username: &newName,
				Email:    &newEmail,
				FullName: &newName,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, pgx.ErrNoRows)
			},
			buildContext: func(t *testing.T, tokenMaker token.TokenMaker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user.ID, user.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.NotFound, st.Code())
			},
		},
		{
			name: "MissingAuthorization",
			req: &pb.UpdateUserRequest{
				Id:       user.ID,
				Username: &newName,
				Email:    &newEmail,
				FullName: &newName,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.TokenMaker) context.Context {
				return context.Background()
			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Unauthenticated, st.Code())
			},
		},
		{
			name: "ExpiredToken",
			req: &pb.UpdateUserRequest{
				Id:       user.ID,
				Username: &newName,
				Email:    &newEmail,
				FullName: &newName,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.TokenMaker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user.ID, user.Role, -time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Unauthenticated, st.Code())
			},
		},
		{
			name: "InvalidUsername",
			req: &pb.UpdateUserRequest{
				Id:       user.ID,
				Username: &invalidUsername,
				Email:    &newEmail,
				FullName: &newName,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.TokenMaker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user.ID, user.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.InvalidArgument, st.Code())
			},
		},
		{
			name: "InvalidEmail",
			req: &pb.UpdateUserRequest{
				Id:       user.ID,
				Username: &newName,
				Email:    &invalidEmail,
				FullName: &newName,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.TokenMaker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user.ID, user.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.InvalidArgument, st.Code())
			},
		},
		{
			name: "InvalidFullName",
			req: &pb.UpdateUserRequest{
				Id:       user.ID,
				Username: &newName,
				Email:    &newEmail,
				FullName: &invalidFullName,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.TokenMaker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user.ID, user.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.InvalidArgument, st.Code())
			},
		},
		{
			name: "InvalidPassword",
			req: &pb.UpdateUserRequest{
				Id:       user.ID,
				Username: &newName,
				Email:    &newEmail,
				FullName: &newName,
				Password: &InvalidPassword,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.TokenMaker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user.ID, user.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.InvalidArgument, st.Code())
			},
		},
		{
			name: "PermissionDenied",
			req: &pb.UpdateUserRequest{
				Id:       user.ID,
				Username: &newName,
				Email:    &newEmail,
				FullName: &newName,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.TokenMaker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, 0, user.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.PermissionDenied, st.Code())
			},
		},
	}

	for _, testCase := range testCases {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		store := mockdb.NewMockStore(ctrl)
		testCase.buildStubs(store)

		server := NewTestServer(t, store)

		res, err := server.UpdateUser(testCase.buildContext(t, server.tokenMaker), testCase.req)
		testCase.checkResponse(t, res, err)
	}
}

func randomUser(t *testing.T) (db.User, string) {
	password := utils.RandomString(8)
	hashedPassword, err := utils.HashPassword(password)
	require.NoError(t, err)
	return db.User{
		ID:             utils.RandomInt(1, 100),
		Username:       utils.RandomName(),
		HashedPassword: hashedPassword,
		FullName:       utils.RandomName(),
		Email:          utils.RandomEmail(),
		Role:           utils.DepositorRole,
	}, password
}
