package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	mocks "github.com/itallix/gophkeeper/mocks/pkg/generated/api/proto/v1"
	pb "github.com/itallix/gophkeeper/pkg/generated/api/proto/v1"
)

func TestCardCommands(t *testing.T) {
	// Save original client and restore after tests
	originalClient := client
	defer func() { client = originalClient }()

	mockClient := mocks.NewGophkeeperServiceClient(t)
	client = mockClient

	t.Run("list cards", func(t *testing.T) {
		buf := new(bytes.Buffer)
		cmd := NewCardCmd()
		cmd.SetOut(buf)

		mockClient.EXPECT().List(mock.Anything, &pb.ListRequest{
			Type: pb.DataType_DATA_TYPE_CARD,
		}).Return(&pb.ListResponse{
			Secrets: []string{"card1", "card2"},
		}, nil)

		cmd.SetArgs([]string{"list"})
		err := cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, buf.String(), "card1")
		assert.Contains(t, buf.String(), "card2")
	})

	t.Run("get card", func(t *testing.T) {
		buf := new(bytes.Buffer)
		cmd := NewCardCmd()
		cmd.SetOut(buf)

		mockClient.EXPECT().Get(mock.Anything, &pb.GetRequest{
			Type: pb.DataType_DATA_TYPE_CARD,
			Path: "test-card",
		}).Return(&pb.GetResponse{
			Data: &pb.TypedData{
				Base: &pb.Metadata{
					Path:      "test-card",
					CreatedAt: "2024-01-01",
					CreatedBy: "user",
					Metadata:  "test-meta",
				},
				Data: &pb.TypedData_Card{
					Card: &pb.CardData{
						CardHolder:  "John Doe",
						Number:      "4111111111111111",
						ExpiryMonth: 12,
						ExpiryYear:  25,
						Cvv:         "123",
					},
				},
			},
		}, nil)

		cmd.SetArgs([]string{"get", "-p", "test-card"})
		err := cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, buf.String(), "John Doe")
		assert.Contains(t, buf.String(), "4111111111111111")
		assert.Contains(t, buf.String(), "12/25")
	})

	t.Run("create card", func(t *testing.T) {
		input := "John Doe\n4111111111111111\n12\n25\n123\n"
		cmd := NewCardCmd()
		cmd.SetIn(strings.NewReader(input))
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)

		expectedReq := &pb.CreateRequest{
			Data: &pb.TypedData{
				Type: pb.DataType_DATA_TYPE_CARD,
				Base: &pb.Metadata{
					Path: "new-card",
				},
				Data: &pb.TypedData_Card{
					Card: &pb.CardData{
						CardHolder:  "John Doe",
						Number:      "4111111111111111",
						ExpiryMonth: 12,
						ExpiryYear:  25,
						Cvv:         "123",
					},
				},
			},
		}

		mockClient.EXPECT().Create(mock.Anything, expectedReq).Return(&pb.CreateResponse{
			Message: "Card created successfully",
		}, nil)

		cmd.SetArgs([]string{"create", "-p", "new-card"})
		err := cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, buf.String(), "Card created successfully")
	})

	t.Run("delete card", func(t *testing.T) {
		buf := new(bytes.Buffer)
		cmd := NewCardCmd()
		cmd.SetOut(buf)

		mockClient.EXPECT().Delete(mock.Anything, &pb.DeleteRequest{
			Type: pb.DataType_DATA_TYPE_CARD,
			Path: "test-card",
		}).Return(&pb.DeleteResponse{
			Message: "Card deleted successfully",
		}, nil)

		cmd.SetArgs([]string{"delete", "-p", "test-card"})
		err := cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, buf.String(), "Card deleted successfully")
	})
}
