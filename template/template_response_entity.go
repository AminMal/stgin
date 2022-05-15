package template

type tmpl struct {
	lines []string
}

func (t tmpl) ContentType() string {
	return "text/html"
}

func (t tmpl) Bytes() ([]byte, error) {
	var bytes []byte
	for _, line := range t.lines {
		bytes = append(bytes, []byte(line)...)
	}
	return bytes, nil
}

