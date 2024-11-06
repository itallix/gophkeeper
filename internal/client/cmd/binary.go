package cmd

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/spf13/cobra"

	pb "gophkeeper.com/pkg/generated/api/proto/v1"
)

const (
	chunkSize = 512 * 1024 // 0.5MB
)

var (
	ErrChunkHash = errors.New("aborted upload due to chunk hash mismatch")
	ErrFileHash  = errors.New("aborted upload due to file hash mismatch")
)

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

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List binaries",
		RunE: func(cmd *cobra.Command, _ []string) error {
			resp, err := client.List(context.Background(), &pb.ListRequest{
				Type: pb.DataType_DATA_TYPE_BINARY,
			})
			if err != nil {
				return fmt.Errorf("error listing binaries: %w", err)
			}
			for _, name := range resp.GetSecrets() {
				cmd.Println(name)
			}
			return nil
		},
	}

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Upload a new binary",
		RunE: func(cmd *cobra.Command, _ []string) error {
			fpath, _ := cmd.Flags().GetString("file")

			file, err := os.Open(fpath)
			if err != nil {
				return fmt.Errorf("failed to read a file: %w", err)
			}
			defer file.Close()

			stream, err := client.Upload(context.Background())
			if err != nil {
				return fmt.Errorf("failed to create upload stream: %w", err)
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
					return fmt.Errorf("failed to read file: %w", err)
				}

				chunkData := buffer[:n]
				chunkHash := fileHash.AddChunk(chunkID, chunkData)

				chunk := &pb.Chunk{
					Data:     chunkData,
					Filename: filepath.Base(fpath),
					Hash:     chunkHash,
					ChunkId:  chunkID,
				}

				if err = stream.Send(chunk); err != nil {
					return fmt.Errorf("failed to send chunk: %w", err)
				}
				cmd.Print(".")
				chunkID++
			}
			cmd.Println()

			if err = stream.Send(&pb.Chunk{
				Data:     nil,
				Filename: filepath.Base(fpath),
				Hash:     fileHash.Complete(),
				ChunkId:  chunkID,
			}); err != nil {
				return fmt.Errorf("failed to send chunk: %w", err)
			}

			resp, err := stream.CloseAndRecv()
			if err != nil {
				return fmt.Errorf("failed to receive upload status: %w", err)
			}

			cmd.Println(resp.GetMessage())
			return nil
		},
	}
	createCmd.Flags().StringP("file", "f", "", "Binary filepath")
	_ = createCmd.MarkFlagRequired("file")

	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get binary data",
		RunE: func(cmd *cobra.Command, _ []string) error {
			path, _ := cmd.Flags().GetString("path")
			output, _ := cmd.Flags().GetString("output")

			file, err := os.Create(output)
			if err != nil {
				return fmt.Errorf("failed to create a new file: %w", err)
			}
			defer file.Close()

			stream, err := client.Download(context.Background(), &pb.DownloadRequest{
				Filename: path,
			})
			if err != nil {
				return fmt.Errorf("failed to create download stream: %w", err)
			}

			fileHash := NewFileHash()
			var i = 0
			for {
				chunk, err := stream.Recv()
				if err == io.EOF {
					return nil
				}
				if err != nil {
					_ = os.Remove(file.Name())
					return fmt.Errorf("failed to receive chunk: %w", err)
				}
				if chunk.Data != nil {
					currentHash := fileHash.AddChunk(chunk.GetChunkId(), chunk.GetData())
					if chunk.GetHash() != currentHash {
						_ = os.Remove(file.Name())
						return ErrChunkHash
					}
					_, err = file.Write(chunk.GetData())
					if err != nil {
						_ = os.Remove(file.Name())
						return fmt.Errorf("aborted due to error writing to file: %w", err)
					}
					i++
					cmd.Print(".")
				} else {
					if chunk.GetHash() != fileHash.Complete() {
						_ = os.Remove(file.Name())
						return ErrFileHash
					}
					break
				}
			}
			cmd.Println()
			cmd.Printf("Download binary %s with %d chunks completed.", output, i)
			return nil
		},
	}
	getCmd.Flags().StringP("path", "p", "", "Binary path")
	getCmd.Flags().StringP("output", "o", "", "Output path")
	_ = getCmd.MarkFlagRequired("file")
	_ = getCmd.MarkFlagRequired("output")

	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete binary",
		RunE: func(cmd *cobra.Command, _ []string) error {
			path, _ := cmd.Flags().GetString("path")

			resp, err := client.Delete(context.Background(), &pb.DeleteRequest{
				Type: pb.DataType_DATA_TYPE_BINARY,
				Path: path,
			})
			if err != nil {
				return fmt.Errorf("error deleting binary: %w", err)
			}
			cmd.Println(resp.GetMessage())
			return nil
		},
	}
	deleteCmd.Flags().StringP("path", "p", "", "Binary path")
	_ = deleteCmd.MarkFlagRequired("path")

	binaryCmd.AddCommand(listCmd, createCmd, getCmd, deleteCmd)

	return binaryCmd
}
