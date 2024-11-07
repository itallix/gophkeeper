package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	pb "gophkeeper.com/pkg/generated/api/proto/v1"
)

func NewListCmd(secretName, desc string, dataType pb.DataType) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: desc,
		RunE: func(cmd *cobra.Command, _ []string) error {
			resp, err := client.List(context.Background(), &pb.ListRequest{
				Type: dataType,
			})
			if err != nil {
				return fmt.Errorf("error listing %s: %w", secretName, err)
			}
			for _, name := range resp.GetSecrets() {
				cmd.Println(name)
			}
			return nil
		},
	}
}

func NewDeleteCmd(secretName, desc string, dataType pb.DataType) *cobra.Command {
	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: desc,
		RunE: func(cmd *cobra.Command, _ []string) error {
			path, _ := cmd.Flags().GetString("path")

			resp, err := client.Delete(context.Background(), &pb.DeleteRequest{
				Type: dataType,
				Path: path,
			})
			if err != nil {
				return fmt.Errorf("error deleting %s: %w", secretName, err)
			}
			cmd.Println(resp.GetMessage())
			return nil
		},
	}
	deleteCmd.Flags().StringP("path", "p", "", secretName+" path")
	_ = deleteCmd.MarkFlagRequired("path")

	return deleteCmd
}
