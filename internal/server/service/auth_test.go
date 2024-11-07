package service_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"golang.org/x/crypto/bcrypt"

	"gophkeeper.com/internal/server"
	"gophkeeper.com/internal/server/service"
	"gophkeeper.com/internal/server/storage"

	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const (
	postgresDatabase = "test"
	postgresUser     = "user"
	postgresPassword = "password"
)

type JWTAuthTestSuite struct {
	postgresContainer testcontainers.Container

	suite.Suite
}

func (suite *JWTAuthTestSuite) SetupSuite() {
	ctx := context.Background()
	postgresContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase(postgresDatabase),
		postgres.WithUsername(postgresUser),
		postgres.WithPassword(postgresPassword),
		testcontainers.WithWaitStrategy(
			wait.ForAll(
				wait.ForLog("database system is ready to accept connections"),
				wait.ForListeningPort("5432/tcp"),
			).WithDeadline(1*time.Minute),
		))
	suite.Require().NoError(err)
	suite.postgresContainer = postgresContainer
}

func (suite *JWTAuthTestSuite) TearDownSuite() {
	ctx := context.Background()
	suite.Require().NoError(suite.postgresContainer.Terminate(ctx))
}

func (suite *JWTAuthTestSuite) TestJWTAuth() {
	ctx := context.Background()
	postgresEndpoint, pgErr := suite.postgresContainer.Endpoint(ctx, "")
	suite.Require().NoError(pgErr)
	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", postgresUser, postgresPassword,
		postgresEndpoint, postgresDatabase)
	suite.Require().NoError(server.ApplyMigrations(dsn, "../../../db/migrations"))
	pool, pgErr := pgxpool.New(ctx, dsn)
	suite.Require().NoError(pgErr)
	userRepo := storage.NewUserRepo(pool)
	authService := service.NewJWTAuthService(
		userRepo,
		[]byte("access-secret-key"),
		[]byte("refresh-secret-key"),
		15*time.Minute,
		24*time.Hour,
	)

	givenUsername := "mark"
	givenPassword := "secret"
	hashedPassword, pwdErr := bcrypt.GenerateFromPassword([]byte(givenPassword), bcrypt.DefaultCost)
	suite.Require().NoError(pwdErr)
	suite.Require().NoError(userRepo.CreateUser(ctx, givenUsername, string(hashedPassword)))

	suite.Run("successful token pair generation & refresh", func() {
		tokens, err := authService.GetTokenPair(givenUsername)

		suite.Require().NoError(err)
		suite.NotEmpty(tokens.AccessToken)
		suite.NotEmpty(tokens.RefreshToken)

		actual, err := authService.ValidateAccessToken(tokens.AccessToken)
		suite.Require().NoError(err)
		suite.Equal(givenUsername, actual)

		claims, err := authService.RefreshTokens(tokens.RefreshToken)
		suite.Require().NoError(err)
		actual, err = authService.ValidateAccessToken(claims.AccessToken)
		suite.Require().NoError(err)
		suite.Equal(givenUsername, actual)
	})

	suite.Run("successful authentication", func() {
		tokens, err := authService.Authenticate(ctx, givenUsername, givenPassword)

		suite.Require().NoError(err)
		suite.NotEmpty(tokens.AccessToken)
		suite.NotEmpty(tokens.RefreshToken)

		actual, err := authService.ValidateAccessToken(tokens.AccessToken)
		suite.Require().NoError(err)
		suite.Equal(givenUsername, actual)
	})

	suite.Run("invalid password", func() {
		_, err := authService.Authenticate(ctx, givenUsername, "geheim")

		suite.Require().Error(err)
		suite.ErrorIs(err, service.ErrInvalidCreds)
	})

	suite.Run("invalid user", func() {
		_, err := authService.Authenticate(ctx, "steve", "geheim")

		suite.Require().Error(err)
		suite.ErrorIs(err, service.ErrUserNotFound)
	})
}

func TestJWTAuthTestSuite(t *testing.T) {
	suite.Run(t, new(JWTAuthTestSuite))
}
