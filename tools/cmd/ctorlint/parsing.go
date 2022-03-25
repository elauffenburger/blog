package main

import (
	"bufio"
	"io"
	"regexp"
	"strings"
	"unicode"
)

var (
	structDeclRegex = regexp.MustCompile(`type (.*?) struct \{`)
	ctorDeclRegex   = regexp.MustCompile(`func (New(.*?))\((?:.*?)\)(.*){`)
)

func getVisibility(str string) visibility {
	var vis visibility
	if unicode.IsUpper(([]rune)(str)[0]) {
		vis = exported
	} else {
		vis = unexported
	}

	return vis
}

func parseAsTyps(str string) ([]typ, error) {
	// All typs we've parsed.
	var ts []typ

	// The current typ we're working on.
	var (
		name      []rune
		ptr       = false
		inGeneric = false
	)
	for _, c := range str {
		if c == ' ' || c == '(' || c == ')' {
			continue
		}

		if len(name) == 0 && c == '*' {
			ptr = true
			continue
		}

		if c == ',' && !inGeneric {
			ts = append(ts, typ{string(name), ptr})

			name = []rune{}
			ptr = false
			inGeneric = false

			continue
		}

		name = append(name, c)
	}

	ts = append(ts, typ{string(name), ptr})

	return ts, nil
}

func findStructs(src io.Reader) ([]strct, error) {
	var strcts []strct

	lines := bufio.NewScanner(src)
	for lines.Scan() {
		l := lines.Text()

		// If this line is a comment asking us to ignore linting for the type on the _next_ line,
		// set nolint to true, advance the scanner to the next line, and update `l`.
		nolint := false
		if strings.HasPrefix(l, "//nolint:ctors") {
			nolint = true
			if !lines.Scan() {
				break
			}

			l = lines.Text()
		}

		matches := structDeclRegex.FindAllStringSubmatch(l, -1)
		for _, mg := range matches {
			if len(mg) < 2 {
				continue
			}

			structName := mg[1]
			strcts = append(strcts, strct{structName, getVisibility(structName), nolint})
		}
	}

	return strcts, nil
}

func findCtors(src string) ([]ctor, error) {
	var ctors []ctor

	matches := ctorDeclRegex.FindAllStringSubmatch(src, -1)
	for _, mg := range matches {
		if len(mg) < 4 {
			continue
		}

		name, constructs, returnsStr := mg[1], mg[2], mg[3]

		var returns []typ
		if len(returnsStr) != 0 {
			typs, err := parseAsTyps(returnsStr)
			if err != nil {
				return nil, err
			}

			returns = typs
		}

		ctors = append(ctors, ctor{
			name,
			constructs,
			getVisibility(name),
			returns,
		})
	}

	return ctors, nil

}
