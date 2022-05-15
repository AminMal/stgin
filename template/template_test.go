package template

import "testing"

var sampleHtmlTemplateFile []string = []string{
	"<p>{{ title   }}</p>",
	"<p>This is {{ name }} talking</p>",
}

func TestTemplateLoadingContents(t *testing.T) {
	lines := loadTemplateContents(sampleHtmlTemplateFile, Variables{
		"title": "STGIN",
		"name": "John Doe",
	})
	if lines[0] != "<p>STGIN</p>" {
		t.Fatal("Template loader could not load html template")
	}
	if lines[1] != "<p>This is John Doe talking</p>" {
		t.Fatal("Template loader could not load html template")
	}
}