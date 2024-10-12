package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
)

type ChatGPTResponse struct {
    Choices []struct {
        Text string `json:"text"`
    } `json:"choices"`
}

func GenerateResponse(prompt string) (string, error) {
    apiKey := os.Getenv("CHATGPT_API_KEY")
    url := "https://api.openai.com/v1/engines/davinci-codex/completions"

    requestBody, err := json.Marshal(map[string]interface{}{
        "prompt": prompt,
        "max_tokens": 150,
    })
    if err != nil {
        return "", err
    }

    req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
    if err != nil {
        return "", err
    }
    req.Header.Set("Authorization", "Bearer "+apiKey)
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    var chatgptResponse ChatGPTResponse
    if err := json.NewDecoder(resp.Body).Decode(&chatgptResponse); err != nil {
        return "", err
    }

    return chatgptResponse.Choices[0].Text, nil
}
