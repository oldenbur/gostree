package gostree

import (
	"bytes"
	"encoding/gob"
	"fmt"

	log "github.com/cihub/seelog"
)

// clone returns a deep copy of the subject STree
func (t STree) clone() (STree, error) {

	gob.Register(STree(map[interface{}]interface{}{}))
	gob.Register([]interface{}{})

	var mod bytes.Buffer
	enc := gob.NewEncoder(&mod)
	dec := gob.NewDecoder(&mod)

	err := enc.Encode(t)
	if err != nil {
		return nil, fmt.Errorf("clone Encode error: %v", err)
	}

	var clone STree
	err = dec.Decode(&clone)
	if err != nil {
		return nil, fmt.Errorf("clone Decode error: %v", err)
	}

	return clone, nil
}

func (t STree) SetVal(path string, val interface{}) (STree, error) {

	clone, err := t.clone()
	if err != nil {
		return nil, fmt.Errorf("SetVal clone error: %v", err)
	}

	p, err := ValueOfPath(path)
	if err != nil {
		return nil, fmt.Errorf("SetVal ValueOfPath error: %v", err)
	}

	return clone.setPathVal(p, val)
}

func (t STree) setPathVal(path FieldPath, val interface{}) (STree, error) {

	if path == nil || len(path) < 1 {
		return t, fmt.Errorf("setPathVal called with no path")
	}

	pathKey, pathIdx, err := t.parsePathComponent(path[0])
	if err != nil {
		return t, fmt.Errorf("setPathVal parsePathComponent error: %v", err)
	}

	var tVal interface{}
	if tv, ok := t[pathKey]; !ok {
		return t, fmt.Errorf("setPathVal path component not found: %s", path[0])
	} else {
		tVal = tv
	}

	if len(path) == 1 && pathIdx < 0 {
		t[pathKey] = val
		return t, nil
	}

	log.Debugf("setPathVal(%v) on %v", path[1:], tVal)

	if IsMap(tVal) {
		t[pathKey], err = tVal.(STree).setPathVal(path[1:], val)
		return t, err

	} else if IsSlice(tVal) {
		sVal := tVal.([]interface{})
		if pathIdx < 0 || pathIdx >= len(sVal) {
			return t, fmt.Errorf("setPathVal invalid slice index %d for path %s", pathIdx, path[0])
		}
		if IsMap(sVal[pathIdx]) {
			_, err = sVal[pathIdx].(STree).setPathVal(path[1:], val)
			return t, err
		} else if len(path) == 1 {
			sVal[pathIdx] = val
			return t, nil
		} else {
			return t, fmt.Errorf("setPathVal unable to traverse below slice path component: %s", path[0])
		}

	} else {
		return t, fmt.Errorf("setPathVal unable to traverse below path component: %s", path[0])
	}

}
