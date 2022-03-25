package main

type visibility string

const (
	exported   visibility = "pub"
	unexported            = ""
)

type strct struct {
	name   string
	vis    visibility
	nolint bool
}

type typ struct {
	name string
	ptr  bool
}

type ctor struct {
	name       string
	constructs string
	vis        visibility
	returns    []typ
}

func (c ctor) MatchesStruct(s strct) bool {
	for _, t := range c.returns {
		if t.name == s.name {
			return true
		}
	}

	return false
}
