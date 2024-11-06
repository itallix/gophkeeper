package cmd

import (
	"bufio"
	"context"
	"fmt"

	"github.com/spf13/cobra"

	pb "gophkeeper.com/pkg/generated/api/proto/v1"
)

func NewNoteCmd() *cobra.Command {
	noteCmd := &cobra.Command{
		Use:   "note",
		Short: "Note management commands",
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List available notes",
		RunE: func(cmd *cobra.Command, _ []string) error {
			resp, err := client.List(context.Background(), &pb.ListRequest{
				Type: pb.DataType_DATA_TYPE_NOTE,
			})
			if err != nil {
				return fmt.Errorf("error listing notes: %w", err)
			}
			for _, name := range resp.GetSecrets() {
				cmd.Println(name)
			}
			return nil
		},
	}

	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Retrieve note data by path",
		RunE: func(cmd *cobra.Command, _ []string) error {
			path, _ := cmd.Flags().GetString("path")

			resp, err := client.Get(context.Background(), &pb.GetRequest{
				Type: pb.DataType_DATA_TYPE_NOTE,
				Path: path,
			})
			if err != nil {
				return fmt.Errorf("failed to retrieve note data: %w", err)
			}
			cmd.Printf("Note: %s\n", resp.GetData().GetNote().GetText())
			cmd.Printf("Created at: %s\n", resp.GetData().GetBase().GetCreatedAt())
			cmd.Printf("Created by: %s\n", resp.GetData().GetBase().GetCreatedBy())
			cmd.Printf("Metadata: %s\n", resp.GetData().GetBase().GetMetadata())
			return nil
		},
	}
	getCmd.Flags().StringP("path", "p", "", "Note path")
	_ = getCmd.MarkFlagRequired("path")

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new note",
		RunE: func(cmd *cobra.Command, _ []string) error {
			path, _ := cmd.Flags().GetString("path")
			reader := bufio.NewReader(cmd.InOrStdin())

			// Read password securely
			text, err := promptString(cmd, reader, "Enter note text: ")
			if err != nil {
				return fmt.Errorf("failed to read note text: %w", err)
			}
			cmd.Println()

			resp, err := client.Create(context.Background(), &pb.CreateRequest{
				Data: &pb.TypedData{
					Type: pb.DataType_DATA_TYPE_NOTE,
					Base: &pb.Metadata{
						Path: path,
					},
					Data: &pb.TypedData_Note{
						Note: &pb.NoteData{
							Text: text,
						},
					},
				},
			})
			if err != nil {
				return fmt.Errorf("failed to create a new note: %w", err)
			}
			cmd.Println(resp.GetMessage())
			return nil
		},
	}
	createCmd.Flags().StringP("path", "p", "", "Note path")
	_ = createCmd.MarkFlagRequired("path")

	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete existing note",
		RunE: func(cmd *cobra.Command, _ []string) error {
			path, _ := cmd.Flags().GetString("path")

			resp, err := client.Delete(context.Background(), &pb.DeleteRequest{
				Type: pb.DataType_DATA_TYPE_NOTE,
				Path: path,
			})
			if err != nil {
				return fmt.Errorf("error deleting note: %w", err)
			}
			cmd.Println(resp.GetMessage())
			return nil
		},
	}
	deleteCmd.Flags().StringP("path", "p", "", "Note path")
	_ = deleteCmd.MarkFlagRequired("path")

	noteCmd.AddCommand(listCmd, getCmd, createCmd, deleteCmd)

	return noteCmd
}
