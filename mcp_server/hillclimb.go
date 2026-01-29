package mcp_server

import (
	"context"

	"puzzle_helper/cmd"
)

// HillclimbRequest defines the input for the Hillclimb solve operation.
type HillclimbRequest struct {
	CipherText      string  `json:"cipherText"`
	Generations     int     `json:"generations"`
	Mutations       int     `json:"mutations"`
	RegenAfter      int     `json:"regenAfter"`
	CandidateCount  int     `json:"candidateCount"`
	LocalLookaround int     `json:"localLookaround"`
	NgramSize       int     `json:"ngramSize"`
}

// HillclimbResponse defines the output for the Hillclimb solve operation.
type HillclimbResponse struct {
	Solutions []cmd.HillclimbResult `json:"solutions"`
}

// HillclimbService defines the interface for Hillclimb operations.
type HillclimbService interface {
	Solve(ctx context.Context, req *HillclimbRequest) (*HillclimbResponse, error)
}
