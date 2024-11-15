package cmd

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	pb "github.com/itallix/gophkeeper/pkg/generated/api/proto/v1"
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
	chunks map[int64]string
}

func NewFileHash() *FileHash {
	return &FileHash{
		hash:   sha256.New(),
		chunks: make(map[int64]string),
	}
}

func (fh *FileHash) AddChunk(chunkID int64, data []byte) string {
	fh.mu.Lock()
	defer fh.mu.Unlock()

	chunkHash := sha256.Sum256(data)
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

func newCreateBinaryCmd() *cobra.Command {
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
			chunkID := int64(0)
			for {
				n, readErr := file.Read(buffer)
				if errors.Is(readErr, io.EOF) {
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
	return createCmd
}

// reassembleBinaryChunks reconstructs a binary file from a stream of chunks received via gRPC.
// It creates a new file at the specified path and verifies the integrity of each chunk and the complete file
// using hash checksums.
//
// The function writes chunks sequentially to the created file while maintaining a running hash.
// If any error occurs during the process, it automatically cleans up by removing the incomplete file.
// Progress is indicated by printing dots to the command output.
//
// Parameters:
//   - filename: Path where the new file will be created
//   - stream: gRPC stream providing ordered chunks of binary data
//   - cmd: Cobra command instance for progress output
//
// Returns:
//   - error: nil on successful reassembly, otherwise:
//   - Wrapped error if file creation fails
//   - ErrChunkHash if a chunk's hash verification fails
//   - ErrFileHash if the complete file's hash verification fails
//   - Wrapped error for I/O or stream reception failures
//
// Example:
//
//	err := reassembleBinaryChunks("output.bin", stream, cmd)
//	if err != nil {
//	    log.Printf("Failed to reassemble file: %v", err)
//	}
func reassembleBinaryChunks(filename string, stream grpc.ServerStreamingClient[pb.Chunk],
	cmd *cobra.Command) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create a new file: %w", err)
	}
	defer file.Close()
	fileHash := NewFileHash()
	var i = 0
	for {
		chunk, recvErr := stream.Recv()
		if errors.Is(recvErr, io.EOF) {
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
	return nil
}

func newGetBinaryCmd() *cobra.Command {
	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get binary data",
		RunE: func(cmd *cobra.Command, _ []string) error {
			path, _ := cmd.Flags().GetString("path")
			output, _ := cmd.Flags().GetString("output")

			stream, err := client.Download(context.Background(), &pb.DownloadRequest{
				Filename: path,
			})
			if err != nil {
				return fmt.Errorf("failed to create download stream: %w", err)
			}

			if err = reassembleBinaryChunks(output, stream, cmd); err != nil {
				return err
			}
			cmd.Println()
			cmd.Printf("Binary %s has been successfully retrieved.", output)
			return nil
		},
	}
	getCmd.Flags().StringP("path", "p", "", "Binary path")
	getCmd.Flags().StringP("output", "o", "", "Output path")
	_ = getCmd.MarkFlagRequired("file")
	_ = getCmd.MarkFlagRequired("output")
	return getCmd
}

func NewBinaryCmd() *cobra.Command {
	binaryCmd := &cobra.Command{
		Use:   "binary",
		Short: "Binary management commands",
	}

	binaryCmd.AddCommand(
		NewListCmd("binary", "List binaries", pb.DataType_DATA_TYPE_BINARY),
		newCreateBinaryCmd(), newGetBinaryCmd(),
		NewDeleteCmd("binary", "Delete binary", pb.DataType_DATA_TYPE_BINARY))

	return binaryCmd
}
