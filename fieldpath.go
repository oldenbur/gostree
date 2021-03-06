package gostree

import (
	"fmt"
	"regexp"
	"strings"
	//	log "github.com/cihub/seelog"
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

func PathString(path ...string) string {
	result := []string{}
	for _, p := range path {
		result = append(result, fmt.Sprintf(".%s", strings.Replace(p, `.`, `\.`, -1)))
	}
	return strings.Join(result, "")
}

func (p FieldPath) shift() FieldPath {
	if len(p) < 1 {
		return p
	} else {
		return p[1:]
	}
}

func (p FieldPath) append(keys ...string) FieldPath {
	for _, key := range keys {
		p = append(p, key)
	}
	return p
}

func (p FieldPath) first() string {
	if len(p) < 1 {
		return ""
	} else {
		return p[0]
	}
}

func (p FieldPath) last() string {
	if len(p) < 1 {
		return ""
	} else {
		return p[len(p)-1]
	}
}

func ValueOfPath(p string) (FieldPath, error) {
	var result []string
	if len(p) < 1 {
		return result, nil
	}
	if !strings.HasPrefix(p, ".") {
		return nil, fmt.Errorf("ValueOfPath lacks prefix .: %s", p)
	}
	subs := pathRegexp.FindAllStringSubmatch(p, -1)
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

func AsPath(keys ...string) string {
	return FieldPath(keys).String()
}

// FieldPaths returns a slice of FieldPaths representing the list of full key paths to
// each "leaf" of the STree.
func (s STree) FieldPaths() (paths []FieldPath) {
	return s.fieldPaths([]string{}, paths)
}

func (s STree) fieldPaths(parent FieldPath, tally []FieldPath) []FieldPath {
	for k, v := range s {

		var f string
		var path FieldPath
		if fVal, ok := k.(string); ok {
			f = fVal
		} else {
			panic(fmt.Sprintf("fieldPaths failed to convert STree k '%v' to Field", k))
		}

		if !IsSlice(v) {
			path = append(parent, f)
		} else {
			tally = fieldPathsSlice(parent, tally, f, v)
			continue
		}

		if vs, err := ValueOf(v); err == nil {
			tally = vs.fieldPaths(path, tally)
		} else {
			tally = append(tally, path)
		}
	}
	return tally
}

func fieldPathsSlice(parent FieldPath, tally []FieldPath, key string, val interface{}) []FieldPath {
	for i, vi := range val.([]interface{}) {
		keySub := fmt.Sprintf("%s[%d]", key, i)
		path := append(parent, keySub)
		if vs, err := ValueOf(vi); err == nil {
			tally = vs.fieldPaths(path, tally)
		} else if IsSlice(vi) {
			tally = fieldPathsSlice(parent, tally, keySub, vi)
		} else {
			tally = append(tally, path)
		}
	}
	return tally
}
