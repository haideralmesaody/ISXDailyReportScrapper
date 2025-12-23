package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type OperationRequest struct {
	OperationType string                 `json:"operation_type"`
	Config        map[string]interface{} `json:"config"`
}

type OperationResponse struct {
	ID      string `json:"id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

func main() {
	// Create the indicators operation request
	req := OperationRequest{
		OperationType: "indicators",
		Config: map[string]interface{}{
			"rsi_period":   14,
			"sma_periods":  "20,50,200",
			"ema_periods":  "20",
			"macd_params":  "12,26,9",
		},
	}

	// Convert to JSON
	jsonData, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("Error marshaling request: %v\n", err)
		return
	}

	// Create HTTP request
	httpReq, err := http.NewRequest("POST", "http://localhost:8080/api/operations", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	fmt.Println("Submitting indicators operation...")

	// Send request
	resp, err := client.Do(httpReq)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		return
	}

	fmt.Printf("Response Status: %s\n", resp.Status)
	fmt.Printf("Response Body: %s\n", string(body))

	if resp.StatusCode == 200 || resp.StatusCode == 201 || resp.StatusCode == 202 {
		fmt.Println("✅ Indicators operation submitted successfully!")

		// Try to parse response if it's JSON
		var opResp OperationResponse
		if err := json.Unmarshal(body, &opResp); err == nil {
			fmt.Printf("Operation ID: %s\n", opResp.ID)
			fmt.Printf("Operation Status: %s\n", opResp.Status)
		}
	} else {
		fmt.Printf("❌ Failed to submit operation. Status: %d\n", resp.StatusCode)
	}
}