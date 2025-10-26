package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/uptrace/bun"

	servicepkgtesting "github.com/clevertechware/todo-bun-app/internal/pkg/testing"
)

func TestPGRepository(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(PGRepositorySuite))
}

type PGRepositorySuite struct {
	suite.Suite
	pgContainer *servicepkgtesting.PostgresTestDatabase
}

func (s *PGRepositorySuite) SetupSuite() {
	req := s.Require()

	// Initialize container
	pgContainer := servicepkgtesting.NewPostgresDatabase()
	s.pgContainer = pgContainer

	s.T().Cleanup(func() {
		req.NoError(s.pgContainer.TearDown())
	})
}

func (s *PGRepositorySuite) insert(t *testing.T, client bun.IDB, seed interface{}) {
	t.Helper()
	_, err := client.NewInsert().Model(seed).Exec(context.Background())
	require.NoError(t, err)
}
