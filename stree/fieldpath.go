package stree

import (
	"fmt"
	"regexp"
	"strings"
)

type FieldPath []string

var pathRegexp *regexp.Regexp = regexp.MustCompile(`\.((?:[^\.\\]|\\\.)+)`)

func (p FieldPath) String() string {
	if len(p) < 1 {
		return ""
	}

	pEsc := "." + strings.Replace(p[0], `.`, `\.`, -1)
	if len(p) == 1 {
		return pEsc
	} else {
		return pEsc + p.shift().String()
	}
}

func (p FieldPath) shift() FieldPath {
	if len(p) < 1 {
		return p
	} else {
		return p[1:]
	}
}

func (p FieldPath) next() string {
	if len(p) < 1 {
		return ""
	} else {
		return p[0]
	}
}

func ValueOfPath(p string) (FieldPath, error) {
	if !strings.HasPrefix(p, ".") {
		return nil, fmt.Errorf("ValueOfPath lacks prefix .: %s", p)
	}
	subs := pathRegexp.FindAllStringSubmatch(p, -1)
	var result []string
	for i, sub := range subs {
		if len(sub) < 2 {
			return result, fmt.Errorf("ValueOfPath(\"%s\") unexpected submatch %d: %q", p, i, subs)
		}
		result = append(result, strings.Replace(sub[1], `\.`, `.`, -1))
	}
	return result, nil
}

func ValueOfPathMust(p string) FieldPath {
	f, err := ValueOfPath(p)
	if err != nil {
		panic(err)
	}
	return f
}

// FieldPaths returns a slice of FieldPaths representing the list of full key paths to
// each "leaf" of the STree.
func (s STree) FieldPaths() (paths []FieldPath) {
	return s.fieldPaths([]string{}, paths)
}

func (s STree) fieldPaths(parent FieldPath, tally []FieldPath) (paths []FieldPath) {
	for k, v := range s {
		var path FieldPath
		if f, ok := k.(string); ok {
			path = append(parent, f)
		} else {
			panic(fmt.Sprintf("fieldPaths failed to convert STree k '%v' to Field", k))
		}

		if vs, err := ValueOf(v); err == nil {
			tally = vs.fieldPaths(path, tally)
		} else {
			tally = append(tally, path)
		}
	}
	return tally
}
