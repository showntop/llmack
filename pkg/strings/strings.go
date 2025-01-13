package strings

import "strings"

// TrimSpecial trim tab \n \r \t 等特殊字符
func TrimSpecial(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\t", " ")
	return s
}
