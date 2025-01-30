package api

import (
	"testing"

	"github.com/stretchr/testify/require"
	db "github.com/valkyraycho/bank_project/db/sqlc"
	"github.com/valkyraycho/bank_project/utils"
)

func NewTestServer(t *testing.T, store db.Store) *Server {
	server, err := NewServer(utils.Config{TokenSymmetricKey: utils.RandomString(32)}, store)
	require.NoError(t, err)
	return server
}
