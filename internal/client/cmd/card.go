package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	pb "gophkeeper.com/pkg/generated/api/proto/v1"
)

func NewCardCmd() *cobra.Command {
	cardCmd := &cobra.Command{
		Use:   "card",
		Short: "Card management commands",
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List available cards",
		Run: func(cmd *cobra.Command, _ []string) {
			resp, err := client.List(context.Background(), &pb.ListRequest{
				Type: pb.DataType_DATA_TYPE_CARD,
			})
			if err != nil {
				fmt.Printf("Error listing cards: %v\n", err)
				os.Exit(1)
			}
			for _, name := range resp.GetSecrets() {
				fmt.Println(name)
			}
		},
	}

	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Retrieve card data by path",
		Run: func(cmd *cobra.Command, _ []string) {
			path, _ := cmd.Flags().GetString("path")

			resp, err := client.Get(context.Background(), &pb.GetRequest{
				Type: pb.DataType_DATA_TYPE_CARD,
				Path: path,
			})
			if err != nil {
				fmt.Printf("Failed to retrieve login data: %v\n", err)
				os.Exit(1)
			}
			baseData := resp.GetData().GetBase()
			cardData := resp.GetData().GetCard()
			fmt.Printf("Card holder: %s\n", cardData.GetCardHolder())
			fmt.Printf("Card number: %s\n", cardData.GetNumber())
			fmt.Printf("Expiry month/year: %d/%d\n", cardData.GetExpiryMonth(), cardData.GetExpiryYear())
			fmt.Printf("CVC: %s\n", cardData.GetCvv())
			fmt.Printf("Created at: %s\n", baseData.GetCreatedAt())
			fmt.Printf("Created by: %s\n", baseData.GetCreatedBy())
			fmt.Printf("Metadata: %s\n", baseData.GetMetadata())
		},
	}
	getCmd.Flags().StringP("path", "p", "", "Card path")
	_ = getCmd.MarkFlagRequired("path")

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new card",
		Run: func(cmd *cobra.Command, _ []string) {
			path, _ := cmd.Flags().GetString("path")

			holderName, err := promptString("Enter card holder name: ")
			if err != nil {
				fmt.Printf("\nFailed to read card number: %v\n", err)
				os.Exit(1)
			}
			number, err := promptString("Enter card number: ")
			if err != nil {
				fmt.Printf("\nFailed to read card number: %v\n", err)
				os.Exit(1)
			}
			expiryMonth, err := promptNumber("Enter expiry month: ")
			if err != nil {
				fmt.Printf("\nFailed to read expiry month: %v\n", err)
				os.Exit(1)
			}
			expiryYear, err := promptNumber("Enter expiry year: ")
			if err != nil {
				fmt.Printf("\nFailed to read expiry year: %v\n", err)
				os.Exit(1)
			}
			fmt.Print("Enter CVC: ")
			cvc, err := term.ReadPassword(int(os.Stdin.Fd()))
			if err != nil {
				fmt.Printf("\nFailed to read CVC: %v\n", err)
				os.Exit(1)
			}
			fmt.Println()

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
							ExpiryMonth: int32(expiryMonth),
							ExpiryYear:  int32(expiryYear),
							Cvv:         string(cvc),
						},
					},
				},
			})
			if err != nil {
				fmt.Printf("Failed to create a new card: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(resp.GetMessage())
		},
	}
	createCmd.Flags().StringP("path", "p", "", "Card path")
	_ = createCmd.MarkFlagRequired("path")

	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete existing card",
		Run: func(cmd *cobra.Command, _ []string) {
			path, _ := cmd.Flags().GetString("path")

			resp, err := client.Delete(context.Background(), &pb.DeleteRequest{
				Type: pb.DataType_DATA_TYPE_CARD,
				Path: path,
			})
			if err != nil {
				fmt.Printf("Error deleting card: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(resp.GetMessage())
		},
	}
	deleteCmd.Flags().StringP("path", "p", "", "Card path")
	_ = deleteCmd.MarkFlagRequired("path")

	cardCmd.AddCommand(listCmd, getCmd, createCmd, deleteCmd)

	return cardCmd
}
