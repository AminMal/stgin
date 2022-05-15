package template

import "regexp"

var templateVariableDefinitionRegex *regexp.Regexp

func init() {
	templateVariableDefinitionRegex = regexp.MustCompile("\\{\\{\\s*([a-zA-Z0-9_-]+)\\s*\\}\\}")
}
