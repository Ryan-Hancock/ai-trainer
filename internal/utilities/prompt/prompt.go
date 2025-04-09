package prompts

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/template"
)

var namedPrompts = make(map[string]string)

func LoadPromptsFromFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Split on `-- name: Name`
	regex := regexp.MustCompile(`(?m)^-- name: (.+)\n`)
	indexes := regex.FindAllStringSubmatchIndex(string(content), -1)

	for i := range indexes {
		name := string(content[indexes[i][2]:indexes[i][3]])
		start := indexes[i][1]
		end := len(content)
		if i+1 < len(indexes) {
			end = indexes[i+1][0]
		}
		namedPrompts[name] = strings.TrimSpace(string(content[start:end]))
	}

	return nil
}

func GetPrompt(name string, data any) (string, error) {
	raw, ok := namedPrompts[name]
	if !ok {
		return "", fmt.Errorf("prompt %s not found", name)
	}

	tmpl, err := template.New(name).Parse(raw)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	err = tmpl.Execute(&b, data)
	return b.String(), err
}
