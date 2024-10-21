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

// CrawlDirectoryWithExt crawls the directory and returns a list of files with the specified extension.
func CrawlDirectoryWithExt(basePath string, ext string) ([]string, error) {
	var files []string
	err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ext {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}
// Run generates documentation based on the provided prompt file and file extension.
func Run(promptFile string, fileExt string, cfg config.Config) error {
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

	files, err := CrawlDirectoryWithExt(currentDir, fileExt)
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
			log.Printf("Failed to get response from Replicate for %s: %v", filePath, err)
			continue
		}

		// Create a relative path for S3
		relativePath, err := filepath.Rel(currentDir, filePath)
		if err != nil {
			log.Printf("Failed to get relative path for %s: %v", filePath, err)
			continue
		}


		// Try to upload the response to S3
		err = s3Service.UploadFile(context.Background(), relativePath+".md", []byte(response))
		if err != nil {
			log.Printf("Failed to upload file to S3: %v. Attempting to save locally...", err)

			// Define the local directory and file path
			localDir := "/tmp/data/"
			localFilePath := filepath.Join(localDir, relativePath+".md")

			// Create the local directory if it doesn't exist
			if err := os.(localDir, os.ModePerm); err != nil {
				log.Printf("Failed to create local directory %s: %v", localDir, err)
				continue
			}

			// Write the response to a local file
			if err := ioutil.WriteFile(localFilePath, []byte(response), 0644); err != nil {
				log.Printf("Failed to write response to local file %s: %v", localFilePath, err)
			} else {
				log.Printf("Response saved to local file %s", localFilePath)
			}
		} else {
			log.Printf("Successfully uploaded %s to S3", relativePath+".md")
		}
	}

	return nil
}
