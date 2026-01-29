package mcp_server

import "context"

// CaesarRequest defines the input for the Caesar cipher operation.
type CaesarRequest struct {
	Text string `json:"text"`
}

// CaesarShiftResult represents a single shifted string.
type CaesarShiftResult struct {
	ShiftedText string `json:"shiftedText"`
	Shift       int    `json:"shift"`
}

// CaesarResponse defines the output for the Caesar cipher operation.
type CaesarResponse struct {
	Shifts []CaesarShiftResult `json:"shifts"`
}

// CaesarService defines the interface for Caesar cipher operations.
type CaesarService interface {
	Shift(ctx context.Context, req *CaesarRequest) (*CaesarResponse, error)
}
