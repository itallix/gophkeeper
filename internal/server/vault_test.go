package server_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	mp "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	m "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/minio"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"gophkeeper.com/internal/server"
	"gophkeeper.com/internal/server/models"
	"gophkeeper.com/internal/server/s3"
	"gophkeeper.com/internal/server/service"
	"gophkeeper.com/internal/server/storage"

	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const (
	minioAccessKey   = "superadmin"
	minioSecretKey   = "superpassword"
	postgresDatabase = "test"
	postgresUser     = "user"
	postgresPassword = "password"
)

type VaultTestSuite struct {
	minioContainer    testcontainers.Container
	postgresContainer testcontainers.Container

	suite.Suite
}

func (suite *VaultTestSuite) SetupSuite() {
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
	require.NoError(suite.T(), err)
	suite.postgresContainer = postgresContainer

	minioContainer, err := minio.Run(ctx,
		"minio/minio:RELEASE.2024-08-03T04-33-23Z",
		minio.WithUsername(minioAccessKey),
		minio.WithPassword(minioSecretKey),
		testcontainers.WithEnv(map[string]string{"MINIO_DEFAULT_BUCKETS": "binaries"}))
	require.NoError(suite.T(), err)

	endpoint, err := minioContainer.Endpoint(ctx, "")
	require.NoError(suite.T(), err)

	client, err := m.New(endpoint, &m.Options{
		Creds:  credentials.NewStaticV4(minioAccessKey, minioSecretKey, ""),
		Secure: false,
	})
	require.NoError(suite.T(), err)

	for _, bucket := range []string{"binaries"} {
		_ = client.MakeBucket(ctx, bucket, m.MakeBucketOptions{})
	}

	suite.minioContainer = minioContainer
}

func (suite *VaultTestSuite) TearDownSuite() {
	ctx := context.Background()
	require.NoError(suite.T(), suite.minioContainer.Terminate(ctx))
	require.NoError(suite.T(), suite.postgresContainer.Terminate(ctx))
}

func (suite *VaultTestSuite) applyMigrations(dsn string) {
	db, err := sql.Open("postgres", dsn)
	require.NoError(suite.T(), err)
	defer db.Close()

	driver, err := mp.WithInstance(db, &mp.Config{})
	require.NoError(suite.T(), err)

	m, err := migrate.NewWithDatabaseInstance(
		"file://../../db/migrations",
		"postgres",
		driver,
	)
	require.NoError(suite.T(), err)

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		require.NoError(suite.T(), err)
	}
}

func (suite *VaultTestSuite) TestVaultAPI() {
	ctx := context.Background()
	postgresEndpoint, err := suite.postgresContainer.Endpoint(ctx, "")
	require.NoError(suite.T(), err)
	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", postgresUser, postgresPassword,
		postgresEndpoint, postgresDatabase)
	suite.applyMigrations(dsn)
	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(suite.T(), err)
	objectStorage, err := s3.NewObjectStorage()
	require.NoError(suite.T(), err)
	kms, err := service.NewRSAKMS()
	require.NoError(suite.T(), err)
	encryptionService := service.NewStandardEncryptionService(kms)
	vault := server.NewVault(ctx, pool, objectStorage, encryptionService)

	userRepo := storage.NewUserRepo(pool)
	username := "mark"
	require.NoError(suite.T(), userRepo.CreateUser(ctx, username, "aurelius"))

	suite.Run("logins", func() {
		secret := models.NewLogin([]models.SecretOption{
			models.WithPath("login0"),
			models.WithCreatedBy(username),
			models.WithModifiedBy(username),
		}, []models.LoginOption{
			models.WithLogin("leo"),
			models.WithPassword("secret"),
		})
		require.NoError(suite.T(), vault.StoreSecret(secret))

		retrieved := models.NewLogin([]models.SecretOption{
			models.WithPath("login0"),
		}, nil)
		require.NoError(suite.T(), vault.RetrieveSecret(retrieved))
		assert.Equal(suite.T(), "leo", retrieved.Login)
		assert.Equal(suite.T(), "secret", string(retrieved.Password))

		secrets, err := vault.ListSecrets(models.NewLogin(nil, nil))
		require.NoError(suite.T(), err)
		assert.Len(suite.T(), secrets, 1)

		deleted := models.NewLogin([]models.SecretOption{
			models.WithPath("login0"),
		}, nil)
		require.NoError(suite.T(), vault.DeleteSecret(deleted))

		secrets, err = vault.ListSecrets(models.NewLogin(nil, nil))
		require.NoError(suite.T(), err)
		assert.Empty(suite.T(), secrets)
	})

	suite.Run("cards", func() {
		secret := models.NewCard([]models.SecretOption{
			models.WithPath("card0"),
			models.WithCreatedBy(username),
			models.WithModifiedBy(username),
		}, []models.CardOption{
			models.WithCardNumber("1122334455667788"),
			models.WithCardHolder("Mark Aurelius"),
			models.WithCVC("247"),
			models.WithExpiry(8, int16(time.Now().Year() + 2)),
		})
		require.NoError(suite.T(), vault.StoreSecret(secret))

		retrieved := models.NewCard([]models.SecretOption{
			models.WithPath("card0"),
		}, nil)
		require.NoError(suite.T(), vault.RetrieveSecret(retrieved))
		assert.Equal(suite.T(), "1122334455667788", string(retrieved.Number))
		assert.Equal(suite.T(), "Mark Aurelius", retrieved.CardholderName)
		assert.Equal(suite.T(), "247", string(retrieved.CVC))
		assert.Equal(suite.T(), int8(8), retrieved.ExpiryMonth)
		assert.Equal(suite.T(), int16(time.Now().Year() + 2), retrieved.ExpiryYear)

		secrets, err := vault.ListSecrets(models.NewCard(nil, nil))
		require.NoError(suite.T(), err)
		assert.Len(suite.T(), secrets, 1)

		deleted := models.NewCard([]models.SecretOption{
			models.WithPath("card0"),
		}, nil)
		require.NoError(suite.T(), vault.DeleteSecret(deleted))

		secrets, err = vault.ListSecrets(models.NewCard(nil, nil))
		require.NoError(suite.T(), err)
		assert.Empty(suite.T(), secrets)
	})

	suite.Run("notes", func() {
		secret := models.NewNote([]models.SecretOption{
			models.WithPath("note0"),
			models.WithCreatedBy(username),
			models.WithModifiedBy(username),
		}, []models.NoteOption{
			models.WithText("lorem ipsum"),
		})
		require.NoError(suite.T(), vault.StoreSecret(secret))

		retrieved := models.NewNote([]models.SecretOption{
			models.WithPath("note0"),
		}, nil)
		require.NoError(suite.T(), vault.RetrieveSecret(retrieved))
		assert.Equal(suite.T(), "lorem ipsum", string(retrieved.Text))

		secrets, err := vault.ListSecrets(models.NewNote(nil, nil))
		require.NoError(suite.T(), err)
		assert.Len(suite.T(), secrets, 1)

		deleted := models.NewNote([]models.SecretOption{
			models.WithPath("note0"),
		}, nil)
		require.NoError(suite.T(), vault.DeleteSecret(deleted))

		secrets, err = vault.ListSecrets(models.NewNote(nil, nil))
		require.NoError(suite.T(), err)
		assert.Empty(suite.T(), secrets)
	})
}

func TestVaultTestSuite(t *testing.T) {
	suite.Run(t, new(VaultTestSuite))
}
