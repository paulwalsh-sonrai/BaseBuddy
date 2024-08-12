package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

const (
	repoOwner        = "sonraisecurity"
	repoName         = "findings-service"
	githubAPI        = "https://api.github.com"
	commitsEndpoint  = "/repos/%s/%s/commits"
	commitEndpoint   = "/repos/%s/%s/commits/%s"
	contentsEndpoint = "/repos/%s/%s/contents/%s"
	timeFormat       = "2006-01-02T15:04:05Z"
	chatPrompt       = "generate simple mermaid diagram code from the code given. Only respond with the mermaid code and no explanation or big comments:\n\n"
)

type Commit struct {
	SHA    string     `json:"sha"`
	Commit CommitInfo `json:"commit"`
}

type CommitInfo struct {
	Author AuthorInfo `json:"author"`
}

type AuthorInfo struct {
	Date string `json:"date"`
}

type CommitDetail struct {
	Files []ChangedFile `json:"files"`
}

type ChangedFile struct {
	FileName string `json:"filename"`
}

type FileContent struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Sha         string `json:"sha"`
	Size        int    `json:"size"`
	Url         string `json:"url"`
	HtmlUrl     string `json:"html_url"`
	GitUrl      string `json:"git_url"`
	DownloadUrl string `json:"download_url"`
	Type        string `json:"type"`
	Content     string `json:"content"`
	Encoding    string `json:"encoding"`
}

type GitHubClient struct {
	baseURL string
	token   string
	client  *http.Client
}

type OpenAIClient struct {
	client *openai.Client
}

func NewGitHubClient() *GitHubClient {
	return &GitHubClient{
		baseURL: githubAPI,
		token:   os.Getenv("GITHUB_TOKEN"),
		client:  &http.Client{},
	}
}

func NewOpenAIClient() *OpenAIClient {
	return &OpenAIClient{
		client: openai.NewClient(os.Getenv("OPENAI_TOKEN")),
	}
}

func (o *OpenAIClient) CreateChatCompletion(ctx context.Context, message string) (string, error) {
	resp, err := o.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT4o,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: chatPrompt + message,
				},
			},
		},
	)
	if err != nil {
		return "", fmt.Errorf("chat completion error: %v", err)
	}
	response := resp.Choices[0].Message.Content
	fmt.Printf("Response from chat: %s", response)
	return response, nil
}

func (c *GitHubClient) GetRecentCommits(owner, repo string) ([]Commit, error) {
	url := fmt.Sprintf(c.baseURL+commitsEndpoint, owner, repo)
	fmt.Printf("Request URL: %s\n", url) // for debugging
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("token %s", c.token))
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch commits: %s", resp.Status)
	}

	var commits []Commit
	err = json.NewDecoder(resp.Body).Decode(&commits)
	if err != nil {
		return nil, err
	}

	return commits, nil
}

func (c *GitHubClient) GetCommitDetails(owner, repo, sha string) (*CommitDetail, error) {
	url := fmt.Sprintf(c.baseURL+commitEndpoint, owner, repo, sha)
	fmt.Printf("Request URL: %s\n", url) // for debugging

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("token %s", c.token))
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch commit details: %s", resp.Status)
	}

	var commitDetail CommitDetail
	err = json.NewDecoder(resp.Body).Decode(&commitDetail)
	if err != nil {
		return nil, err
	}

	return &commitDetail, nil
}

func (c *GitHubClient) GetFileContent(owner, repo, path string) (*FileContent, error) {
	url := fmt.Sprintf(c.baseURL+contentsEndpoint, owner, repo, path)
	fmt.Printf("Request URL: %s\n", url) // for debugging
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("token %s", c.token))
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// It is possible the file that was modified was deleted, we may want to handle this diffferently later on
	if resp.StatusCode == http.StatusNotFound {
		fmt.Printf("File %s not found (likely deleted)\n", path)
		return nil, fmt.Errorf("file not found: %s", path) // Return nil to indicate that the file was not found
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch file: %s\n request: %s", resp.Status, url)
	}

	var fileContent FileContent
	err = json.NewDecoder(resp.Body).Decode(&fileContent)
	if err != nil {
		return nil, err
	}

	return &fileContent, nil
}

// this will be replaced with a batch put to S3
func saveFile(filePath string, data []byte) error {
	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		return err
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	return err
}

// to map files changed to SHA commit
func (c *GitHubClient) FetchFilesMap(owner, repo string) (map[string]string, error) {
	commits, err := c.GetRecentCommits(owner, repo)
	if err != nil {
		return nil, err
	}

	filesMap := make(map[string]string)

	now := time.Now()
	twentyFourHoursAgo := now.Add(-70 * time.Hour)

	for _, commit := range commits {
		commitTime, err := time.Parse(time.RFC3339, commit.Commit.Author.Date)
		if err != nil {
			return nil, fmt.Errorf("error parsing commit date: %v", err)
		}
		if commitTime.After(twentyFourHoursAgo) {
			commitDetail, err := c.GetCommitDetails(owner, repo, commit.SHA)
			if err != nil {
				return nil, err
			}
			for _, file := range commitDetail.Files {
				fileContent, err := c.GetFileContent(owner, repo, file.FileName)
				if err != nil {
					if err.Error() == fmt.Sprintf("file not found: %s", file.FileName) {
						// skip file if possibly deleted
						continue
					}
					return nil, fmt.Errorf("error fetching file content for %s: %v", file.FileName, err)
				}
				// Decode base64 content if necessary
				var content []byte
				if fileContent.Encoding == "base64" {
					content, err = base64.StdEncoding.DecodeString(fileContent.Content)
					if err != nil {
						return nil, fmt.Errorf("error decoding base64 content for %s: %v", file.FileName, err)
					}
				} else {
					content = []byte(fileContent.Content)
				}
				filesMap[file.FileName] = string(content)
			}
		}
	}
	fmt.Printf("filesMap: %v", filesMap)
	return filesMap, nil
}

func (o *OpenAIClient) FetchMermaidCode(content string) (string, error) {
	mermaidCode, err := o.CreateChatCompletion(context.Background(), content)
	if err != nil {
		return "", fmt.Errorf("error generating Mermaid code: %v", err)
	}
	return mermaidCode, nil
}

func FetchMermaidCodeMap(client *GitHubClient, openAIClient *OpenAIClient, filesMap map[string]string, baseDir string) error {
	for fileName, content := range filesMap {

		// Save raw code, there is definitely a better way to save the raw code, just going to put it here for now
		rawFilePath := filepath.Join(baseDir, "raw_files", fileName)
		err := saveFile(rawFilePath, []byte(content))

		if err != nil {
			return fmt.Errorf("error saving raw file %s: %v", rawFilePath, err)
		}

		// Generate Mermaid code
		mermaidCode, err := openAIClient.FetchMermaidCode(content)
		if err != nil {
			return fmt.Errorf("error generating Mermaid code: %v", err)
		}

		// Determine the file path to save and save as .md file for now
		filePath := filepath.Join(baseDir, "mermaid_files", fileName+".md")

		// Save Mermaid code locally for now, this will be sent to S3
		err = saveFile(filePath, []byte(mermaidCode))
		if err != nil {
			return fmt.Errorf("error saving file %s: %v", filePath, err)
		}

		fmt.Printf("Generated Mermaid code for %s:\n%s\n", fileName, mermaidCode)
	}
	return nil
}

func main() {
	client := NewGitHubClient()
	openAIClient := NewOpenAIClient()

	filesMap, err := client.FetchFilesMap(repoOwner, repoName)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	err = FetchMermaidCodeMap(client, openAIClient, filesMap, "downloaded_files")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Files fetched and saved successfully")
}
