package server_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	m "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
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
	suite.Require().NoError(err)
	suite.postgresContainer = postgresContainer

	minioContainer, err := minio.Run(ctx,
		"minio/minio:RELEASE.2024-08-03T04-33-23Z",
		minio.WithUsername(minioAccessKey),
		minio.WithPassword(minioSecretKey),
		testcontainers.WithEnv(map[string]string{"MINIO_DEFAULT_BUCKETS": "binaries"}))
	suite.Require().NoError(err)

	endpoint, err := minioContainer.Endpoint(ctx, "")
	suite.Require().NoError(err)

	client, err := m.New(endpoint, &m.Options{
		Creds:  credentials.NewStaticV4(minioAccessKey, minioSecretKey, ""),
		Secure: false,
	})
	suite.Require().NoError(err)

	for _, bucket := range []string{"binaries"} {
		_ = client.MakeBucket(ctx, bucket, m.MakeBucketOptions{})
	}

	suite.minioContainer = minioContainer
}

func (suite *VaultTestSuite) TearDownSuite() {
	ctx := context.Background()
	suite.Require().NoError(suite.minioContainer.Terminate(ctx))
	suite.Require().NoError(suite.postgresContainer.Terminate(ctx))
}

func (suite *VaultTestSuite) TestVaultAPI() {
	ctx := context.Background()
	postgresEndpoint, err := suite.postgresContainer.Endpoint(ctx, "")
	suite.Require().NoError(err)
	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", postgresUser, postgresPassword,
		postgresEndpoint, postgresDatabase)
	suite.Require().NoError(server.ApplyMigrations(dsn, "../../db/migrations"))
	pool, err := pgxpool.New(ctx, dsn)
	suite.Require().NoError(err)
	minioEndpoint, err := suite.minioContainer.Endpoint(ctx, "")
	suite.Require().NoError(err)
	suite.T().Setenv("ENDPOINT", minioEndpoint)
	objectStorage, err := s3.NewObjectStorage()
	suite.Require().NoError(err)
	kms, err := service.NewRSAKMS("../../testdata/private.pem", "../../testdata/encrypted_key.bin")
	suite.Require().NoError(err)
	encryptionService := service.NewStandardEncryptionService(kms)
	vault := server.NewVault(ctx, pool, objectStorage, encryptionService)

	userRepo := storage.NewUserRepo(pool)
	username := "mark"
	suite.Require().NoError(userRepo.CreateUser(ctx, username, "aurelius"))

	suite.Run("logins", func() {
		secret := models.NewLogin([]models.SecretOption{
			models.WithPath("login0"),
			models.WithCreatedBy(username),
			models.WithModifiedBy(username),
		}, []models.LoginOption{
			models.WithLogin("leo"),
			models.WithPassword("secret"),
		})
		suite.Require().NoError(vault.StoreSecret(secret))

		retrieved := models.NewLogin([]models.SecretOption{
			models.WithPath("login0"),
		}, nil)
		suite.Require().NoError(vault.RetrieveSecret(retrieved))
		suite.Equal("leo", retrieved.Login)
		suite.Equal("secret", string(retrieved.Password))

		var secrets []string
		secrets, err = vault.ListSecrets(models.NewLogin(nil, nil))
		suite.Require().NoError(err)
		suite.Len(secrets, 1)

		deleted := models.NewLogin([]models.SecretOption{
			models.WithPath("login0"),
		}, nil)
		suite.Require().NoError(vault.DeleteSecret(deleted))

		secrets, err = vault.ListSecrets(models.NewLogin(nil, nil))
		suite.Require().NoError(err)
		suite.Empty(secrets)
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
			models.WithExpiry(8, int64(time.Now().Year()+2)),
		})
		suite.Require().NoError(vault.StoreSecret(secret))

		retrieved := models.NewCard([]models.SecretOption{
			models.WithPath("card0"),
		}, nil)
		suite.Require().NoError(vault.RetrieveSecret(retrieved))
		suite.Equal("1122334455667788", string(retrieved.Number))
		suite.Equal("Mark Aurelius", retrieved.CardholderName)
		suite.Equal("247", string(retrieved.CVC))
		suite.Equal(int64(8), retrieved.ExpiryMonth)
		suite.Equal(int64(time.Now().Year()+2), retrieved.ExpiryYear)

		var secrets []string
		secrets, err = vault.ListSecrets(models.NewCard(nil, nil))
		suite.Require().NoError(err)
		suite.Len(secrets, 1)

		deleted := models.NewCard([]models.SecretOption{
			models.WithPath("card0"),
		}, nil)
		suite.Require().NoError(vault.DeleteSecret(deleted))

		secrets, err = vault.ListSecrets(models.NewCard(nil, nil))
		suite.Require().NoError(err)
		suite.Empty(secrets)
	})

	suite.Run("notes", func() {
		secret := models.NewNote([]models.SecretOption{
			models.WithPath("note0"),
			models.WithCreatedBy(username),
			models.WithModifiedBy(username),
		}, []models.NoteOption{
			models.WithText("lorem ipsum"),
		})
		suite.Require().NoError(vault.StoreSecret(secret))

		retrieved := models.NewNote([]models.SecretOption{
			models.WithPath("note0"),
		}, nil)
		suite.Require().NoError(vault.RetrieveSecret(retrieved))
		suite.Equal("lorem ipsum", string(retrieved.Text))

		var secrets []string
		secrets, err = vault.ListSecrets(models.NewNote(nil, nil))
		suite.Require().NoError(err)
		suite.Len(secrets, 1)

		deleted := models.NewNote([]models.SecretOption{
			models.WithPath("note0"),
		}, nil)
		suite.Require().NoError(vault.DeleteSecret(deleted))

		secrets, err = vault.ListSecrets(models.NewNote(nil, nil))
		suite.Require().NoError(err)
		suite.Empty(secrets)
	})

	suite.Run("binaries", func() {
		calcHash := func(data []byte) string {
			dataHash := sha256.Sum256(data)
			return hex.EncodeToString(dataHash[:])
		}

		chunk := models.NewBinary([]models.SecretOption{
			models.WithPath("binary0"),
			models.WithCreatedBy(username),
			models.WithModifiedBy(username),
		}, []models.BinaryOption{
			models.WithChunkID(0),
			models.WithData([]byte("test data")),
			models.WithHash(calcHash([]byte("test data"))),
		})
		suite.Require().NoError(vault.StoreSecret(chunk))
		chunk = models.NewBinary([]models.SecretOption{
			models.WithPath("binary0"),
			models.WithCreatedBy(username),
			models.WithModifiedBy(username),
			models.WithEncryptedDataKey(chunk.EncryptedDataKey),
		}, []models.BinaryOption{
			models.WithChunks(1),
			models.WithHash(calcHash([]byte("test data"))),
		})
		suite.Require().NoError(vault.StoreSecret(chunk))

		retrieved := models.NewBinary([]models.SecretOption{
			models.WithPath("binary0"),
		}, nil)
		suite.Require().NoError(vault.RetrieveSecret(retrieved))
		suite.Equal(int64(1), retrieved.Chunks)
		retrieved = models.NewBinary([]models.SecretOption{
			models.WithPath("binary0"),
			models.WithEncryptedDataKey(retrieved.EncryptedDataKey),
		}, []models.BinaryOption{
			models.WithChunks(retrieved.Chunks),
		})
		suite.Require().NoError(vault.RetrieveSecret(retrieved))
		suite.Equal(retrieved.Data, []byte("test data"))

		var secrets []string
		secrets, err = vault.ListSecrets(models.NewBinary(nil, nil))
		suite.Require().NoError(err)
		suite.Len(secrets, 1)

		deleted := models.NewBinary([]models.SecretOption{
			models.WithPath("binary0"),
		}, nil)
		suite.Require().NoError(vault.DeleteSecret(deleted))

		secrets, err = vault.ListSecrets(models.NewBinary(nil, nil))
		suite.Require().NoError(err)
		suite.Empty(secrets)
	})
}

func TestVaultTestSuite(t *testing.T) {
	suite.Run(t, new(VaultTestSuite))
}
