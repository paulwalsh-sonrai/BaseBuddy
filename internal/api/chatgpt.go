package api

import (
	"context"
	"fmt"

	replicate "github.com/replicate/replicate-go"
)

// GenerateResponse generates a response from the LLaMA model based on the provided prompt.
func GenerateResponse(prompt string) (string, error) {
	// Initialize the Replicate client
	r8, err := replicate.NewClient(replicate.WithTokenFromEnv())
	if err != nil {
		return "", fmt.Errorf("failed to create replicate client: %w", err)
	}

	// Set the model and version
	model := "meta/meta-llama-3-70b-instruct"

	// Prepare the input for the model
	input := replicate.PredictionInput{
		"prompt": prompt,
	}

	// Run the model and wait for its output
	output, err := r8.Run(context.TODO(), model, input, nil)
	if err != nil {
		return "", fmt.Errorf("failed to run model: %w", err)
	}
	// Handle the output correctly
	if response, ok := output.([]interface{}); ok && len(response) > 0 {
		tmp := []string{}
		for _, r := range response {
			tmp = append(tmp, r.(string))
		}

		return strings.join("")
	}

}
