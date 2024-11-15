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

func TestNoteCommands(t *testing.T) {
	// Save original client and restore after tests
	originalClient := client
	defer func() { client = originalClient }()

	mockClient := mocks.NewGophkeeperServiceClient(t)
	client = mockClient

	t.Run("list notes", func(t *testing.T) {
		buf := new(bytes.Buffer)
		cmd := NewNoteCmd()
		cmd.SetOut(buf)

		mockClient.EXPECT().List(mock.Anything, &pb.ListRequest{
			Type: pb.DataType_DATA_TYPE_NOTE,
		}).Return(&pb.ListResponse{
			Secrets: []string{"note1", "note2"},
		}, nil)

		cmd.SetArgs([]string{"list"})
		err := cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, buf.String(), "note1")
		assert.Contains(t, buf.String(), "note2")
	})

	t.Run("get note", func(t *testing.T) {
		buf := new(bytes.Buffer)
		cmd := NewNoteCmd()
		cmd.SetOut(buf)

		mockClient.EXPECT().Get(mock.Anything, &pb.GetRequest{
			Type: pb.DataType_DATA_TYPE_NOTE,
			Path: "test-note",
		}).Return(&pb.GetResponse{
			Data: &pb.TypedData{
				Base: &pb.Metadata{
					Path:      "test-note",
					CreatedAt: "2024-01-01",
					CreatedBy: "user",
					Metadata:  "test-meta",
				},
				Data: &pb.TypedData_Note{
					Note: &pb.NoteData{
						Text: "lorem ipsum",
					},
				},
			},
		}, nil)

		cmd.SetArgs([]string{"get", "-p", "test-note"})
		err := cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, buf.String(), "lorem ipsum")
	})

	t.Run("create note", func(t *testing.T) {
		input := "lorem ipsum\n"
		cmd := NewNoteCmd()
		cmd.SetIn(strings.NewReader(input))
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)

		expectedReq := &pb.CreateRequest{
			Data: &pb.TypedData{
				Type: pb.DataType_DATA_TYPE_NOTE,
				Base: &pb.Metadata{
					Path: "new-note",
				},
				Data: &pb.TypedData_Note{
					Note: &pb.NoteData{
						Text: "lorem ipsum",
					},
				},
			},
		}

		mockClient.EXPECT().Create(mock.Anything, expectedReq).Return(&pb.CreateResponse{
			Message: "Note created successfully",
		}, nil)

		cmd.SetArgs([]string{"create", "-p", "new-note"})
		err := cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, buf.String(), "Note created successfully")
	})

	t.Run("delete note", func(t *testing.T) {
		buf := new(bytes.Buffer)
		cmd := NewNoteCmd()
		cmd.SetOut(buf)

		mockClient.EXPECT().Delete(mock.Anything, &pb.DeleteRequest{
			Type: pb.DataType_DATA_TYPE_NOTE,
			Path: "test-note",
		}).Return(&pb.DeleteResponse{
			Message: "Note deleted successfully",
		}, nil)

		cmd.SetArgs([]string{"delete", "-p", "test-note"})
		err := cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, buf.String(), "Note deleted successfully")
	})
}
