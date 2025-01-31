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

func TestCreateTransfer(t *testing.T) {
	fromUser, fromAccount := randomAccount(t)
	_, toAccount := randomAccount(t)

	amount := utils.RandomInt(1, 100)

	transfer := db.Transfer{
		ID:            utils.RandomInt(1, 100),
		FromAccountID: fromAccount.ID,
		ToAccountID:   toAccount.ID,
		Amount:        amount,
		CreatedAt:     time.Now(),
	}

	fromEntry := db.Entry{
		ID:        utils.RandomInt(1, 100),
		AccountID: fromAccount.ID,
		Amount:    -amount,
		CreatedAt: time.Now(),
	}

	toEntry := db.Entry{
		ID:        utils.RandomInt(1, 100),
		AccountID: toAccount.ID,
		Amount:    amount,
		CreatedAt: time.Now(),
	}

	testCases := []struct {
		name          string
		req           *pb.CreateTransferRequest
		buildStubs    func(store *mockdb.MockStore)
		buildContext  func(t *testing.T, tokenMaker token.TokenMaker) context.Context
		checkResponse func(t *testing.T, res *pb.CreateTransferResponse, err error)
	}{
		{
			name: "OK",
			req: &pb.CreateTransferRequest{
				FromAccountId: fromAccount.ID,
				ToAccountId:   toAccount.ID,
				Amount:        amount,
				Currency:      utils.CAD,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).
					Times(1).
					Return(fromAccount, nil)

				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(toAccount.ID)).
					Times(1).
					Return(toAccount, nil)

				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Eq(db.TransferTxParams{
						FromAccountID: fromAccount.ID,
						ToAccountID:   toAccount.ID,
						Amount:        amount,
					})).
					Times(1).
					Return(db.TransferTxResult{
						Transfer:    transfer,
						FromAccount: fromAccount,
						ToAccount:   toAccount,
						FromEntry:   fromEntry,
						ToEntry:     toEntry,
					}, nil)
			},
			buildContext: func(t *testing.T, tokenMaker token.TokenMaker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, fromUser.ID, fromUser.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.CreateTransferResponse, err error) {
				require.NoError(t, err)
				require.NotEmpty(t, res)
				createdTransfer := res.GetTransfer()
				require.Equal(t, transfer.FromAccountID, createdTransfer.FromAccountId)
				require.Equal(t, transfer.ToAccountID, createdTransfer.ToAccountId)
				require.Equal(t, transfer.Amount, createdTransfer.Amount)
			},
		},
		{
			name: "InternalErrorFromAccount",
			req: &pb.CreateTransferRequest{
				FromAccountId: fromAccount.ID,
				ToAccountId:   toAccount.ID,
				Amount:        amount,
				Currency:      utils.CAD,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)

				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)

				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.TokenMaker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, fromUser.ID, fromUser.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.CreateTransferResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Internal, st.Code())
			},
		},
		{
			name: "InternalErrorToAccount",
			req: &pb.CreateTransferRequest{
				FromAccountId: fromAccount.ID,
				ToAccountId:   toAccount.ID,
				Amount:        amount,
				Currency:      utils.CAD,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(1).
					Return(fromAccount, nil)

				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)

				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.TokenMaker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, fromUser.ID, fromUser.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.CreateTransferResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Internal, st.Code())
			},
		},
		{
			name: "InternalErrorTransfer",
			req: &pb.CreateTransferRequest{
				FromAccountId: fromAccount.ID,
				ToAccountId:   toAccount.ID,
				Amount:        amount,
				Currency:      utils.CAD,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(1).
					Return(fromAccount, nil)

				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(1).
					Return(toAccount, nil)

				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.TransferTxResult{}, sql.ErrConnDone)
			},
			buildContext: func(t *testing.T, tokenMaker token.TokenMaker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, fromUser.ID, fromUser.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.CreateTransferResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Internal, st.Code())
			},
		},
		{
			name: "ExpiredToken",
			req: &pb.CreateTransferRequest{
				FromAccountId: fromAccount.ID,
				ToAccountId:   toAccount.ID,
				Amount:        amount,
				Currency:      utils.CAD,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)

				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)

				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.TokenMaker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, fromUser.ID, fromUser.Role, -time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.CreateTransferResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Unauthenticated, st.Code())
			},
		},
		{
			name: "MissingAuthorization",
			req: &pb.CreateTransferRequest{
				FromAccountId: fromAccount.ID,
				ToAccountId:   toAccount.ID,
				Amount:        amount,
				Currency:      utils.CAD,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)

				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)

				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.TokenMaker) context.Context {
				return context.Background()
			},
			checkResponse: func(t *testing.T, res *pb.CreateTransferResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Unauthenticated, st.Code())
			},
		},
		{
			name: "PermissionDenied",
			req: &pb.CreateTransferRequest{
				FromAccountId: fromAccount.ID,
				ToAccountId:   toAccount.ID,
				Amount:        amount,
				Currency:      utils.CAD,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).
					Times(1).
					Return(fromAccount, nil)

				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)

				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.TokenMaker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, 0, fromUser.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.CreateTransferResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.PermissionDenied, st.Code())
			},
		},
		{
			name: "FromAccountNotFound",
			req: &pb.CreateTransferRequest{
				FromAccountId: fromAccount.ID,
				ToAccountId:   toAccount.ID,
				Amount:        amount,
				Currency:      utils.CAD,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).
					Times(1).
					Return(db.Account{}, pgx.ErrNoRows)

				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)

				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.TokenMaker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, fromUser.ID, fromUser.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.CreateTransferResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.NotFound, st.Code())
			},
		},
		{
			name: "ToAccountNotFound",
			req: &pb.CreateTransferRequest{
				FromAccountId: fromAccount.ID,
				ToAccountId:   toAccount.ID,
				Amount:        amount,
				Currency:      utils.CAD,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).
					Times(1).
					Return(fromAccount, nil)

				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(toAccount.ID)).
					Times(1).
					Return(db.Account{}, pgx.ErrNoRows)

				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.TokenMaker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, fromUser.ID, fromUser.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.CreateTransferResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.NotFound, st.Code())
			},
		},
		{
			name: "SameAccountID",
			req: &pb.CreateTransferRequest{
				FromAccountId: fromAccount.ID,
				ToAccountId:   fromAccount.ID,
				Amount:        amount,
				Currency:      utils.CAD,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)

				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)

				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.TokenMaker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, fromUser.ID, fromUser.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.CreateTransferResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.InvalidArgument, st.Code())
			},
		},
		{
			name: "UnmatchedCurrency",
			req: &pb.CreateTransferRequest{
				FromAccountId: fromAccount.ID,
				ToAccountId:   toAccount.ID,
				Amount:        amount,
				Currency:      utils.EUR,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).
					Times(1).
					Return(fromAccount, nil)

				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(toAccount.ID)).
					Times(1).
					Return(toAccount, nil)

				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.TokenMaker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, fromUser.ID, fromUser.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.CreateTransferResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.FailedPrecondition, st.Code())
			},
		},
		{
			name: "InvalidFromAccountID",
			req: &pb.CreateTransferRequest{
				FromAccountId: -1,
				ToAccountId:   toAccount.ID,
				Amount:        amount,
				Currency:      utils.CAD,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)

				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)

				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.TokenMaker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, fromUser.ID, fromUser.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.CreateTransferResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.InvalidArgument, st.Code())
			},
		},
		{
			name: "InvalidToAccountID",
			req: &pb.CreateTransferRequest{
				FromAccountId: fromAccount.ID,
				ToAccountId:   -1,
				Amount:        amount,
				Currency:      utils.CAD,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)

				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)

				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.TokenMaker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, fromUser.ID, fromUser.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.CreateTransferResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.InvalidArgument, st.Code())
			},
		},
		{
			name: "InvalidAmount",
			req: &pb.CreateTransferRequest{
				FromAccountId: fromAccount.ID,
				ToAccountId:   toAccount.ID,
				Amount:        -1,
				Currency:      utils.CAD,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)

				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)

				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.TokenMaker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, fromUser.ID, fromUser.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.CreateTransferResponse, err error) {
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
		res, err := server.CreateTransfer(testCase.buildContext(t, server.tokenMaker), testCase.req)
		testCase.checkResponse(t, res, err)
	}
}
