package utils

import (
	"fmt"
)

func GeneratePrompt(template string, code string) string {
    return fmt.Sprintf(template, code)
}
