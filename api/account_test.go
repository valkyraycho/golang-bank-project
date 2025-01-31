package api

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
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

func TestCreateAccount(t *testing.T) {
	user, account := randomAccount(t)

	testCases := []struct {
		name          string
		req           *pb.CreateAccountRequest
		buildStubs    func(store *mockdb.MockStore)
		buildContext  func(t *testing.T, tokenMaker token.TokenMaker) context.Context
		checkResponse func(t *testing.T, res *pb.CreateAccountResponse, err error)
	}{
		{
			name: "OK",
			req: &pb.CreateAccountRequest{
				OwnerId:  user.ID,
				Currency: account.Currency,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Eq(db.CreateAccountParams{
						OwnerID:  user.ID,
						Currency: account.Currency,
						Balance:  0,
					})).
					Times(1).
					Return(db.Account{
						ID:        1,
						OwnerID:   user.ID,
						Balance:   0,
						Currency:  account.Currency,
						CreatedAt: time.Now(),
					}, nil)
			},
			buildContext: func(t *testing.T, tokenMaker token.TokenMaker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user.ID, user.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.CreateAccountResponse, err error) {
				require.NoError(t, err)
				require.NotEmpty(t, res)
				createdAccount := res.GetAccount()
				require.NotEmpty(t, createdAccount)
				require.Equal(t, user.ID, createdAccount.OwnerId)
				require.Equal(t, account.Balance, createdAccount.Balance)
				require.Equal(t, account.Currency, createdAccount.Currency)
			},
		},
		{
			name: "InternalError",
			req: &pb.CreateAccountRequest{
				OwnerId:  user.ID,
				Currency: account.Currency,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)
			},
			buildContext: func(t *testing.T, tokenMaker token.TokenMaker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user.ID, user.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.CreateAccountResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Internal, st.Code())
			},
		},
		{
			name: "ExpiredToken",
			req: &pb.CreateAccountRequest{
				OwnerId:  user.ID,
				Currency: account.Currency,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.TokenMaker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user.ID, user.Role, -time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.CreateAccountResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Unauthenticated, st.Code())
			},
		},
		{
			name: "MissingAuthorization",
			req: &pb.CreateAccountRequest{
				OwnerId:  user.ID,
				Currency: account.Currency,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.TokenMaker) context.Context {
				return context.Background()
			},
			checkResponse: func(t *testing.T, res *pb.CreateAccountResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Unauthenticated, st.Code())
			},
		},
		{
			name: "PermissionDenied",
			req: &pb.CreateAccountRequest{
				OwnerId:  user.ID,
				Currency: account.Currency,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.TokenMaker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, 0, user.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.CreateAccountResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.PermissionDenied, st.Code())
			},
		},
		{
			name: "UserNotFound",
			req: &pb.CreateAccountRequest{
				OwnerId:  user.ID,
				Currency: account.Currency,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Account{}, pgx.ErrNoRows)
			},
			buildContext: func(t *testing.T, tokenMaker token.TokenMaker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user.ID, user.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.CreateAccountResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.NotFound, st.Code())
			},
		},
		{
			name: "AccountAlreadyExists",
			req: &pb.CreateAccountRequest{
				OwnerId:  user.ID,
				Currency: account.Currency,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Account{}, db.ErrUniqueViolation)
			},
			buildContext: func(t *testing.T, tokenMaker token.TokenMaker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user.ID, user.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.CreateAccountResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.AlreadyExists, st.Code())
			},
		},
		{
			name: "InvalidID",
			req: &pb.CreateAccountRequest{
				OwnerId:  -1,
				Currency: account.Currency,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.TokenMaker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user.ID, user.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.CreateAccountResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.InvalidArgument, st.Code())
			},
		},
		{
			name: "InvalidCurrency",
			req: &pb.CreateAccountRequest{
				OwnerId:  user.ID,
				Currency: "invalid",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.TokenMaker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user.ID, user.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.CreateAccountResponse, err error) {
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
		res, err := server.CreateAccount(testCase.buildContext(t, server.tokenMaker), testCase.req)
		testCase.checkResponse(t, res, err)
	}
}

func TestGetAccount(t *testing.T) {
	user, account := randomAccount(t)

	testCases := []struct {
		name          string
		req           *pb.GetAccountRequest
		buildStubs    func(store *mockdb.MockStore)
		buildContext  func(t *testing.T, tokenMaker token.TokenMaker) context.Context
		checkResponse func(t *testing.T, res *pb.GetAccountResponse, err error)
	}{
		{
			name: "OK",
			req:  &pb.GetAccountRequest{Id: account.ID},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)
			},
			buildContext: func(t *testing.T, tokenMaker token.TokenMaker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user.ID, user.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.GetAccountResponse, err error) {
				require.NoError(t, err)
				require.NotEmpty(t, res)
				gotAccount := res.GetAccount()
				require.NotEmpty(t, gotAccount)
				require.Equal(t, user.ID, gotAccount.OwnerId)
				require.Equal(t, account.Balance, gotAccount.Balance)
				require.Equal(t, account.Currency, gotAccount.Currency)
			},
		},
		{
			name: "InternalError",
			req:  &pb.GetAccountRequest{Id: account.ID},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)
			},
			buildContext: func(t *testing.T, tokenMaker token.TokenMaker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user.ID, user.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.GetAccountResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Internal, st.Code())
			},
		},
		{
			name: "ExpiredToken",
			req:  &pb.GetAccountRequest{Id: account.ID},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.TokenMaker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user.ID, user.Role, -time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.GetAccountResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Unauthenticated, st.Code())
			},
		},
		{
			name: "MissingAuthorization",
			req:  &pb.GetAccountRequest{Id: account.ID},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.TokenMaker) context.Context {
				return context.Background()
			},
			checkResponse: func(t *testing.T, res *pb.GetAccountResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Unauthenticated, st.Code())
			},
		},
		{
			name: "PermissionDenied",
			req:  &pb.GetAccountRequest{Id: account.ID},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)
			},
			buildContext: func(t *testing.T, tokenMaker token.TokenMaker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, 0, user.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.GetAccountResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.PermissionDenied, st.Code())
			},
		},
		{
			name: "AccountNotFound",
			req:  &pb.GetAccountRequest{Id: account.ID},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Account{}, pgx.ErrNoRows)
			},
			buildContext: func(t *testing.T, tokenMaker token.TokenMaker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user.ID, user.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.GetAccountResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.NotFound, st.Code())
			},
		},
		{
			name: "InvalidID",
			req:  &pb.GetAccountRequest{Id: -1},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.TokenMaker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user.ID, user.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.GetAccountResponse, err error) {
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
		res, err := server.GetAccount(testCase.buildContext(t, server.tokenMaker), testCase.req)
		testCase.checkResponse(t, res, err)
	}
}

func TestGetAccounts(t *testing.T) {
	user, _ := randomUser(t)
	accounts := randomAccountsFromUser(user.ID)

	defaultLimit := int32(5)
	defaultOffset := int32(0)

	invalidLimit := int32(-1)
	invalidOffset := int32(-1)

	testCases := []struct {
		name          string
		req           *pb.GetAccountsRequest
		buildStubs    func(store *mockdb.MockStore)
		buildContext  func(t *testing.T, tokenMaker token.TokenMaker) context.Context
		checkResponse func(t *testing.T, res *pb.GetAccountsResponse, err error)
	}{
		{
			name: "OK",
			req:  &pb.GetAccountsRequest{Limit: &defaultLimit, Offset: &defaultOffset},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccount(gomock.Any(), gomock.Eq(db.ListAccountParams{
						OwnerID: user.ID,
						Limit:   defaultLimit,
						Offset:  defaultOffset,
					})).
					Times(1).
					Return(accounts, nil)
			},
			buildContext: func(t *testing.T, tokenMaker token.TokenMaker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user.ID, user.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.GetAccountsResponse, err error) {
				require.NoError(t, err)
				require.NotEmpty(t, res)
				gotAccounts := res.GetAccounts()
				require.NotEmpty(t, gotAccounts)
				require.Len(t, gotAccounts, 3)

				seen := make(map[string]bool)
				for _, account := range gotAccounts {
					require.NotZero(t, account.Id)
					require.Equal(t, user.ID, account.OwnerId)
					require.Equal(t, int32(0), account.Balance)
					require.Contains(t, utils.SupportedCurrencies, account.Currency)
					require.NotContains(t, seen, account.Currency)
					seen[account.Currency] = true
				}
			},
		},
		{
			name: "InternalError",
			req:  &pb.GetAccountsRequest{Limit: &defaultLimit, Offset: &defaultOffset},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccount(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil, sql.ErrConnDone)
			},
			buildContext: func(t *testing.T, tokenMaker token.TokenMaker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user.ID, user.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.GetAccountsResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Internal, st.Code())
			},
		},
		{
			name: "ExpiredToken",
			req:  &pb.GetAccountsRequest{Limit: &defaultLimit, Offset: &defaultOffset},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.TokenMaker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user.ID, user.Role, -time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.GetAccountsResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Unauthenticated, st.Code())
			},
		},
		{
			name: "MissingAuthorization",
			req:  &pb.GetAccountsRequest{Limit: &defaultLimit, Offset: &defaultOffset},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.TokenMaker) context.Context {
				return context.Background()
			},
			checkResponse: func(t *testing.T, res *pb.GetAccountsResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Unauthenticated, st.Code())
			},
		},
		{
			name: "PermissionDenied",
			req:  &pb.GetAccountsRequest{Limit: &defaultLimit, Offset: &defaultOffset},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccount(gomock.Any(), gomock.Any()).
					Times(1).
					Return(accounts, nil)
			},
			buildContext: func(t *testing.T, tokenMaker token.TokenMaker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, -1, user.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.GetAccountsResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.PermissionDenied, st.Code())
			},
		},
		{
			name: "InvalidLimit",
			req:  &pb.GetAccountsRequest{Limit: &invalidLimit, Offset: &defaultOffset},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.TokenMaker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user.ID, user.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.GetAccountsResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.InvalidArgument, st.Code())
			},
		},
		{
			name: "InvalidOffset",
			req:  &pb.GetAccountsRequest{Limit: &defaultLimit, Offset: &invalidOffset},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.TokenMaker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user.ID, user.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.GetAccountsResponse, err error) {
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
		res, err := server.GetAccounts(testCase.buildContext(t, server.tokenMaker), testCase.req)
		testCase.checkResponse(t, res, err)
	}
}

func randomAccount(t *testing.T) (db.User, db.Account) {
	user, _ := randomUser(t)
	return user, db.Account{
		ID:       utils.RandomInt(1, 100),
		OwnerID:  user.ID,
		Balance:  0,
		Currency: utils.CAD,
	}
}

func randomAccountsFromUser(userID int32) []db.Account {
	accounts := []db.Account{}

	for i, currency := range utils.SupportedCurrencies {
		accounts = append(accounts, db.Account{
			ID:       int32(i + 1),
			OwnerID:  userID,
			Currency: currency,
			Balance:  0,
		})
	}

	return accounts
}
