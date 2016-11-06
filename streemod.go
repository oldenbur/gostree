package stree

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

// clone returns a deep copy of the subject STree
func (t STree) clone() (STree, error) {

	gob.Register(STree(map[interface{}]interface{}{}))

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

	return clone, nil
}
