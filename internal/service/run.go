package service

import (
	"basebuddy/internal/api"
	"basebuddy/internal/config"
	"basebuddy/internal/utils"
	"context"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func Run(promptFile string, cfg config.Config ) error {
    // Read the prompt template
    promptTemplate, err := ioutil.ReadFile(promptFile)
    if err != nil {
        return err
    }

    // Create an S3 service instance
    s3Service := NewS3Service(cfg.S3Bucket)

    // Crawl the current directory and subdirectories
    currentDir, err := os.Getwd()
    if err != nil {
        return err
    }

    files, err := utils.CrawlDirectory(currentDir)
    if err != nil {
        return err
    }

    // Iterate through files and generate documentation
    for _, filePath := range files {
        // Read file contents
        code, err := utils.ReadFileContents(filePath)
        if err != nil {
            log.Printf("Failed to read file %s: %v", filePath, err)
            continue
        }

        // Generate the prompt
        prompt := utils.GeneratePrompt(string(promptTemplate), code)

        // Call ChatGPT to get the response
        response, err := api.GenerateResponse(prompt)
        if err != nil {
            log.Printf("Failed to get response from ChatGPT for %s: %v", filePath, err)
            continue
        }

        // Create a relative path for S3
        relativePath, err := filepath.Rel(currentDir, filePath)
        if err != nil {
            log.Printf("Failed to get relative path for %s: %v", filePath, err)
            continue
        }

        // Upload the response to S3
        if err := s3Service.UploadFile(context.Background(), relativePath+".md", []byte(response)); err != nil {
            log.Printf("Failed to upload file to S3: %v", err)
        }
    }

    return nil
}
