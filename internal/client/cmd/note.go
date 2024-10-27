package cmd

import (
	"context"
	"fmt"
	"os"

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
		Run: func(cmd *cobra.Command, _ []string) {
			resp, err := client.List(context.Background(), &pb.ListRequest{
				Type: pb.DataType_DATA_TYPE_NOTE,
			})
			if err != nil {
				fmt.Printf("Error listing notes: %v\n", err)
				os.Exit(1)
			}
			for _, name := range resp.GetSecrets() {
				fmt.Println(name)
			}
		},
	}

	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Retrieve note data by path",
		Run: func(cmd *cobra.Command, _ []string) {
			path, _ := cmd.Flags().GetString("path")

			resp, err := client.Get(context.Background(), &pb.GetRequest{
				Type: pb.DataType_DATA_TYPE_NOTE,
				Path: path,
			})
			if err != nil {
				fmt.Printf("Failed to retrieve note data: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Note: %s\n", resp.GetData().GetNote().Text)
			fmt.Printf("Created at: %s\n", resp.GetData().GetBase().GetCreatedAt())
			fmt.Printf("Created by: %s\n", resp.GetData().GetBase().GetCreatedBy())
			fmt.Printf("Metadata: %s\n", resp.GetData().GetBase().GetMetadata())
		},
	}
	getCmd.Flags().StringP("path", "p", "", "Note path")
	_ = getCmd.MarkFlagRequired("path")

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new note",
		Run: func(cmd *cobra.Command, _ []string) {
			path, _ := cmd.Flags().GetString("path")

			// Read password securely
			text, err := promptString("Enter note text: ")
			if err != nil {
				fmt.Printf("\nFailed to read note text: %v\n", err)
				os.Exit(1)
			}			
			fmt.Println()

			resp, err := client.Create(context.Background(), &pb.CreateRequest{
				Data: &pb.TypedData{
					Type: pb.DataType_DATA_TYPE_NOTE,
					Base: &pb.Metadata{
						Path: path,
					},
					Data: &pb.TypedData_Note{
						Note: &pb.NoteData{
							Text:    text,
						},
					},
				},
			})
			if err != nil {
				fmt.Printf("Failed to create a new note: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(resp.GetMessage())
		},
	}
	createCmd.Flags().StringP("path", "p", "", "Note path")
	_ = createCmd.MarkFlagRequired("path")

	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete existing note",
		Run: func(cmd *cobra.Command, _ []string) {
			path, _ := cmd.Flags().GetString("path")

			resp, err := client.Delete(context.Background(), &pb.DeleteRequest{
				Type: pb.DataType_DATA_TYPE_NOTE,
				Path: path,
			})
			if err != nil {
				fmt.Printf("Error deleting note: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(resp.GetMessage())
		},
	}
	deleteCmd.Flags().StringP("path", "p", "", "Note path")
	_ = deleteCmd.MarkFlagRequired("path")

	noteCmd.AddCommand(listCmd, getCmd, createCmd, deleteCmd)

	return noteCmd
}
