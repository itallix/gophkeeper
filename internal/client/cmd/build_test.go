package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBuildCmd(t *testing.T) {
	tests := []struct {
		name       string
		version    string
		date       string
		commit     string
		wantOutput string
	}{
		{
			name:       "basic build info",
			version:    "v1.0.0",
			date:       "2024-01-01",
			commit:     "abc123",
			wantOutput: "Build version: v1.0.0\nBuild date: 2024-01-01\nBuild commit: abc123",
		},
		{
			name:       "empty fields",
			version:    "",
			date:       "",
			commit:     "",
			wantOutput: "Build version: N/A\nBuild date: N/A\nBuild commit: N/A",
		},
		{
			name:       "special characters",
			version:    "v1.0.0-alpha",
			date:       "2024-01-01T12:00:00Z",
			commit:     "abc123!@#",
			wantOutput: "Build version: v1.0.0-alpha\nBuild date: 2024-01-01T12:00:00Z\nBuild commit: abc123!@#",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)

			cmd := NewBuildCmd(tt.version, tt.date, tt.commit)
			cmd.SetOut(buf)
			err := cmd.Execute()

			require.NoError(t, err)
			assert.Equal(t, tt.wantOutput, buf.String())
		})
	}
}
