package stgin

import (
	"github.com/AminMal/slogger/colored"
	"strings"
	"unicode/utf8"
)

func splitBy(str string, token string) (string, string) {
	arr := strings.SplitN(str, token, 2)
	switch len(arr) {
	case 2:
		return arr[0], arr[1]
	case 1:
		return arr[0], ""
	default:
		return "", ""
	}
}

func getColor(status int) colored.Color {
	switch {
	case status > 100 && status < 300:
		return colored.GREEN
	case status >= 300 && status < 500:
		return colored.YELLOW
	case status >= 500:
		return colored.RED
	default:
		return colored.CYAN
	}
}

func trimFirstRune(s string) string {
	_, i := utf8.DecodeRuneInString(s)
	return s[i:]
}
