package testdal

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/maximmihin/aw25/internal/dal"
	"github.com/stretchr/testify/require"
	"testing"
)

func NewTestDal(t *testing.T, dbConnStr string) *dal.Dal {

	dbPool, err := pgxpool.New(t.Context(), dbConnStr)
	require.NoError(t, err)
	t.Cleanup(dbPool.Close)
	testDal, err := dal.New(t.Context(), dbPool)
	require.NoError(t, err)

	return testDal

}
