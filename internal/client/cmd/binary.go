package cmd

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/spf13/cobra"

	pb "gophkeeper.com/pkg/generated/api/proto/v1"
)

const chunkSize = 1024 * 1024 // 1MB

type FileHash struct {
	mu     sync.Mutex
	hash   hash.Hash
	chunks map[int32]string
}

func NewFileHash() *FileHash {
	return &FileHash{
		hash:   md5.New(),
		chunks: make(map[int32]string),
	}
}

func (fh *FileHash) AddChunk(chunkID int32, data []byte) string {
	fh.mu.Lock()
	defer fh.mu.Unlock()

	chunkHash := md5.Sum(data)
	hexHash := hex.EncodeToString(chunkHash[:])
	fh.chunks[chunkID] = hexHash
	fh.hash.Write(data)

	return hexHash
}

func (fh *FileHash) Complete() string {
	fh.mu.Lock()
	defer fh.mu.Unlock()
	return hex.EncodeToString(fh.hash.Sum(nil))
}

func NewBinaryCmd() *cobra.Command {
	binaryCmd := &cobra.Command{
		Use:   "binary",
		Short: "Binary management commands",
	}

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Upload a new binary",
		Run: func(cmd *cobra.Command, _ []string) {
			fpath, _ := cmd.Flags().GetString("file")

			file, err := os.Open(fpath)
			if err != nil {
				fmt.Printf("Failed to read a file: %v\n", err)
				os.Exit(1)
			}
			defer file.Close()

			stream, err := client.Upload(context.Background())
			if err != nil {
				fmt.Printf("failed to create upload stream: %v\n", err)
				os.Exit(1)
			}

			fileHash := NewFileHash()
			buffer := make([]byte, chunkSize)
			chunkID := int32(0)
			for {
				n, err := file.Read(buffer)
				if err == io.EOF {
					break
				}
				if err != nil {
					fmt.Printf("failed to read file: %v\n", err)
					os.Exit(1)
				}

				chunkData := buffer[:n]
				chunkHash := fileHash.AddChunk(chunkID, chunkData)

				chunk := &pb.Chunk{
					Data:     chunkData,
					Filename: filepath.Base(fpath),
					Hash:     chunkHash,
					ChunkId:  chunkID,
				}

				if err := stream.Send(chunk); err != nil {
					fmt.Printf("failed to send chunk: %v", err)
					os.Exit(1)
				}
				chunkID++
			}

			if err := stream.Send(&pb.Chunk{
				Data:     nil,
				Filename: filepath.Base(fpath),
				Hash:     fileHash.Complete(),
				ChunkId:  chunkID,
			}); err != nil {
				fmt.Printf("failed to send chunk: %v", err)
				os.Exit(1)
			}

			resp, err := stream.CloseAndRecv()
			if err != nil {
				fmt.Printf("Failed to receive upload status: %v\n", err)
				os.Exit(1)
			}

			fmt.Println(resp.GetMessage())
		},
	}
	createCmd.Flags().StringP("file", "f", "", "Binary filepath")
	_ = createCmd.MarkFlagRequired("file")

	binaryCmd.AddCommand(createCmd)

	return binaryCmd
}