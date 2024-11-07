package cmd

import (
	"bufio"
	"context"
	"fmt"

	"github.com/spf13/cobra"

	pb "gophkeeper.com/pkg/generated/api/proto/v1"
)

func NewCardCmd() *cobra.Command {
	cardCmd := &cobra.Command{
		Use:   "card",
		Short: "Card management commands",
	}

	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Retrieve card data by path",
		RunE: func(cmd *cobra.Command, _ []string) error {
			path, _ := cmd.Flags().GetString("path")

			resp, err := client.Get(context.Background(), &pb.GetRequest{
				Type: pb.DataType_DATA_TYPE_CARD,
				Path: path,
			})
			if err != nil {
				return fmt.Errorf("failed to retrieve login data: %w", err)
			}
			baseData := resp.GetData().GetBase()
			cardData := resp.GetData().GetCard()
			cmd.Printf("Card holder: %s\n", cardData.GetCardHolder())
			cmd.Printf("Card number: %s\n", cardData.GetNumber())
			cmd.Printf("Expiry month/year: %d/%d\n", cardData.GetExpiryMonth(), cardData.GetExpiryYear())
			cmd.Printf("CVC: %s\n", cardData.GetCvv())
			cmd.Printf("Created at: %s\n", baseData.GetCreatedAt())
			cmd.Printf("Created by: %s\n", baseData.GetCreatedBy())
			cmd.Printf("Metadata: %s\n", baseData.GetMetadata())
			return nil
		},
	}
	getCmd.Flags().StringP("path", "p", "", "Card path")
	_ = getCmd.MarkFlagRequired("path")

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new card",
		RunE: func(cmd *cobra.Command, _ []string) error {
			path, _ := cmd.Flags().GetString("path")
			reader := bufio.NewReader(cmd.InOrStdin())

			holderName, err := promptString(cmd, reader, "Enter card holder name: ")
			if err != nil {
				return fmt.Errorf("failed to read card number: %w", err)
			}
			number, err := promptString(cmd, reader, "Enter card number: ")
			if err != nil {
				return fmt.Errorf("failed to read card number: %w", err)
			}
			expiryMonth, err := promptNumber(cmd, reader, "Enter expiry month: ")
			if err != nil {
				return fmt.Errorf("failed to read expiry month: %w", err)
			}
			expiryYear, err := promptNumber(cmd, reader, "Enter expiry year: ")
			if err != nil {
				return fmt.Errorf("failed to read expiry year: %w", err)
			}
			cvc, err := promptPassword(cmd, reader, "Enter CVC: ")
			if err != nil {
				return fmt.Errorf("failed to read CVC: %w", err)
			}
			cmd.Println()

			resp, err := client.Create(context.Background(), &pb.CreateRequest{
				Data: &pb.TypedData{
					Type: pb.DataType_DATA_TYPE_CARD,
					Base: &pb.Metadata{
						Path: path,
					},
					Data: &pb.TypedData_Card{
						Card: &pb.CardData{
							CardHolder:  holderName,
							Number:      number,
							ExpiryMonth: int64(expiryMonth),
							ExpiryYear:  int64(expiryYear),
							Cvv:         cvc,
						},
					},
				},
			})
			if err != nil {
				return fmt.Errorf("failed to create a new card: %w", err)
			}
			cmd.Println(resp.GetMessage())
			return nil
		},
	}
	createCmd.Flags().StringP("path", "p", "", "Card path")
	_ = createCmd.MarkFlagRequired("path")

	listCmd := NewListCmd("card", "List available cards", pb.DataType_DATA_TYPE_CARD)
	deleteCmd := NewDeleteCmd("card", "Delete existing card", pb.DataType_DATA_TYPE_CARD)

	cardCmd.AddCommand(listCmd, getCmd, createCmd, deleteCmd)

	return cardCmd
}
