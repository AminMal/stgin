package stgin

import "strings"

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
