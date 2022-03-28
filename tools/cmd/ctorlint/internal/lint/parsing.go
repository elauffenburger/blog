package lint

import (
	"regexp"
	"unicode"
)

var (
	structDeclRegex = regexp.MustCompile(`type (.*?) struct \{`)
	ctorDeclRegex   = regexp.MustCompile(`func (New(.*?))\((?:.*?)\)(.*){`)
)

func getVisibility(name string) Visibility {
	var vis Visibility
	if unicode.IsUpper(([]rune)(name)[0]) {
		vis = Exported
	} else {
		vis = Unexported
	}

	return vis
}
