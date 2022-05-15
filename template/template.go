package template


import (
	"bufio"
	"fmt"
	"github.com/AminMal/stgin"
	"log"
	"os"
	"regexp"
)

type Variables = map[string]string

func loadTemplateRaw(path string) (lines []string, err error) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err = scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return
}

func loadTemplate(path string, variables Variables) ([]string, error) {
	rawLines, err := loadTemplateRaw(path)
	if err != nil {
		return []string{}, err
	}
	var result []string
	for _, line := range rawLines {
		for _, variable := range templateVariableDefinitionRegex.FindAllStringSubmatch(line, -1) {
			variableTitle := variable[1]
			if variableTitle != "" {
				variablePattern := regexp.MustCompile(fmt.Sprintf("\\{\\{\\s*%s\\s*\\}\\}", variableTitle))
				value, exists := variables[variableTitle]
				if exists { line = variablePattern.ReplaceAllString(line, value) }
			}
		}
		result = append(result, line)
	}
	return result, nil
}

func LoadTemplate(path string, variables Variables) stgin.ResponseEntity {
    lines, err := loadTemplate(path, variables)
    if err != nil { panic(err) }
	return tmpl {
		lines: lines,
	}
}