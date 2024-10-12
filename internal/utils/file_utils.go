package utils

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

func CrawlDirectory(basePath string) ([]string, error) {
    var files []string
    err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if !info.IsDir() {
            files = append(files, path)
        }
        return nil
    })
    return files, err
}

func ReadFileContents(filePath string) (string, error) {
    content, err := ioutil.ReadFile(filePath)
    if err != nil {
        return "", err
    }
    return string(content), nil
}
