package api

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	db "github.com/valkyraycho/bank_project/db/sqlc"
	"github.com/valkyraycho/bank_project/token"
	"github.com/valkyraycho/bank_project/utils"
	"google.golang.org/grpc/metadata"
)

func NewTestServer(t *testing.T, store db.Store) *Server {
	server, err := NewServer(utils.Config{TokenSymmetricKey: utils.RandomString(32)}, store)
	require.NoError(t, err)
	return server
}

func newContextWithBearerToken(t *testing.T, tokenMaker token.TokenMaker, user_id int32, role string, duration time.Duration) context.Context {
	accessToken, _, err := tokenMaker.CreateToken(user_id, role, duration)
	require.NoError(t, err)
	return metadata.NewIncomingContext(context.Background(), metadata.MD{
		authorizationHeader: []string{
			fmt.Sprintf("%s %s", authorizationBearer, accessToken),
		},
	})
}
