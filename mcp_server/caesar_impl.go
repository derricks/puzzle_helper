package mcp_server

import (
	"context"
	"encoding/json"
	"net/http"
	"puzzle_helper/cmd"
)

type caesarServiceImpl struct{}

func NewCaesarService() CaesarService {
	return &caesarServiceImpl{}
}

func (s *caesarServiceImpl) Shift(ctx context.Context, req *CaesarRequest) (*CaesarResponse, error) {
	// Call the core logic from the cmd package
	cmdResults := cmd.PerformCaesarShifts(req.Text)

	// Convert cmdResults to MCP CaesarResponse format
	var mcpShifts []CaesarShiftResult
	for _, res := range cmdResults {
		mcpShifts = append(mcpShifts, CaesarShiftResult{
			ShiftedText: res.ShiftedText,
			Shift:       res.Shift,
		})
	}

	return &CaesarResponse{Shifts: mcpShifts}, nil
}

// HandleCaesarShift provides an HTTP handler for the Caesar cipher shift operation.
func HandleCaesarShift(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is supported", http.StatusMethodNotAllowed)
		return
	}

	var req CaesarRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	caesarService := NewCaesarService()
	resp, err := caesarService.Shift(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
