package mcp_server

import (
	"context"

	"puzzle_helper/cmd"
)

// SubstitutionRequest defines the input for the Substitution solve operation.
type SubstitutionRequest struct {
	CipherText           string  `json:"cipherText"`
	NgramSize            int     `json:"ngramSize"`
	Concurrency          int     `json:"concurrency"`
}

// SubstitutionResponse defines the output for the Substitution solve operation.
type SubstitutionResponse struct {
	Solutions []cmd.SubstitutionResult `json:"solutions"`
}

// SubstitutionService defines the interface for Substitution operations.
type SubstitutionService interface {
	Solve(ctx context.Context, req *SubstitutionRequest) (*SubstitutionResponse, error)
}
