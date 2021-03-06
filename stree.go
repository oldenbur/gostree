package gostree

import (
	"bytes"
	"encoding/json"
	"fmt"
	yaml "gopkg.in/yaml.v2"
	"io"
	"reflect"
	"regexp"
	"strconv"

	log "github.com/cihub/seelog"
)

type settingsMap map[string]*reflect.Value

type STree map[interface{}]interface{}

func NewSTree() STree {
	return map[interface{}]interface{}{}
}

func NewSTreeCopy(t STree) (STree, error) {
	return t.clone()
}

// NewSTreeYaml reads yaml from the specified reader, parses it and returns
// the structure as an STree.
func NewSTreeYaml(r io.Reader) (stree STree, err error) {

	buf := bytes.NewBuffer([]byte{})
	_, err = buf.ReadFrom(r)
	if err != nil {
		return nil, fmt.Errorf("NewSTreeJson error reading bytes: %v", err)
	}

	err = yaml.Unmarshal(buf.Bytes(), &stree)
	if err != nil {
		return nil, fmt.Errorf("NewSTreeYaml error in yaml.Unmarshal: ", err)
	}
	return
}

func (s STree) WriteYaml() ([]byte, error) {
	return yaml.Marshal(s)
}

func (s STree) WriteJson(indent bool) ([]byte, error) {

	iMap, err := s.unconvertKeys()
	log.Debugf("iMap: %#v", iMap)
	if err != nil {
		return nil, fmt.Errorf("WriteJson error in unconvertKeys: %v", err)
	}

	var output []byte

	if indent {
		output, err = json.MarshalIndent(iMap, ``, `  `)
	} else {
		output, err = json.Marshal(iMap)
	}

	return output, err
}

// NewSTreeJson reads json from the specified reader, parses it and returns
// the structure as an STree.
func NewSTreeJson(r io.Reader) (stree STree, err error) {

	buf := bytes.NewBuffer([]byte{})
	_, err = buf.ReadFrom(r)
	if err != nil {
		return nil, fmt.Errorf("NewSTreeJson error reading bytes: %v", err)
	}

	um := make(map[string]interface{})
	err = json.Unmarshal(buf.Bytes(), &um)
	if err != nil {
		return nil, fmt.Errorf("NewSTreeJson error in yaml.Unmarshal: ", err)
	}

	return convertKeys(um)
}

func findStructElemsPath(pre string, s interface{}, valsIn settingsMap) (vals settingsMap, err error) {

	vals = valsIn

	v := reflect.ValueOf(s)
	if v.Kind() != reflect.Ptr && v.Kind() != reflect.Interface {
		return vals, fmt.Errorf("findStructElems requires Ptr or Interface, got %s", v.Kind())
	}

	r := v.Elem()
	rType := r.Type()
	for i := 0; i < r.NumField(); i++ {

		f := r.Field(i)

		if isPrimitiveKind(f.Kind()) {
			vals[rType.Field(i).Name] = &f
		}
	}

	return vals, nil
}

// Size returns the number of leaf entries in the STree
// TODO: cache keys?
func (t STree) Size() int {
	return len(t.Keys())
}

// Keys returns a slice containing all top-level keys of this STree
func (t STree) Keys() []interface{} {
	keys := []interface{}{}
	for k, _ := range t {
		keys = append(keys, k)
	}
	return keys
}

// KeyStrings returns a slice containing all top-level keys of this STree converted to
// string, or an error if the conversion fails for any key
func (t STree) KeyStrings() ([]string, error) {
	keys := []string{}
	for k, _ := range t {
		if s, ok := k.(string); !ok {
			return keys, fmt.Errorf("KeyStrings failed to convert %v to string type", k)
		} else {
			keys = append(keys, s)
		}
	}
	return keys, nil
}

// keyRegexp matches strings of the form key_name or slice_name[123]
var keyRegexp *regexp.Regexp = regexp.MustCompile(`^([^\[\]]+)(?:\[(\d+)\])?$`)

// Val returns the leaf value at the position specified by path, which is a slash delimited
// list of nested keys in data, e.g. .level1.level2.key. If the key does not exist, an error
// is returned.
func (t STree) Val(path string) (interface{}, error) {

	keys, err := ValueOfPath(path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse path: %s", path)
	}
	keyCur := keys.first()

	key, idx, err := t.parsePathComponent(keyCur)

	if len(keys) < 1 {
		return nil, fmt.Errorf("no key remaining components")

	} else if len(keys) == 1 && idx < 0 {
		//		log.Debugf("Val(%s) - LastKey: %v", path, t[key])
		if val, ok := t[key]; !ok {
			return nil, fmt.Errorf("Val item not found at key %s", keyCur)
		} else {
			return val, nil
		}

	} else if data, ok := t[key].(STree); ok {
		if idx >= 0 {
			return nil, fmt.Errorf("Val unexpected index for STree value: %s", keyCur)
		} else {
			return data.Val(keys.shift().String())
		}

	} else if data, ok := t[key].([]interface{}); ok {
		// TODO: break this case out to recursively handle nested slices
		//		log.Debugf("Val(%s) - slice: %v", path, data)
		if idx >= 0 && idx < len(data) {
			result := data[idx]
			if len(keys) <= 1 {
				return result, nil
			} else if sval, ok := result.(STree); ok {
				return sval.Val(keys.shift().String())
			}

		} else if idx < 0 {
			if len(keys) > 1 {
				return nil, fmt.Errorf("Val requires index to traverse slice value for key: %s", keyCur)
			}
			return data, nil

		} else {
			return nil, fmt.Errorf("Val slice key index out of range [0,%d]: %s", len(data)-1, keyCur)
		}
	}

	return nil, fmt.Errorf("Val failed to produce value for key: %s", keyCur)
}

func (t STree) ValMust(path string) interface{} {
	val, err := t.Val(path)
	if err != nil {
		panic(err)
	}
	return val
}

// SVal returns the value stored in data at the path, converting it
// to a a string, and returning the zero value if the string is not
// found.
func (t STree) StrVal(path string) (string, error) {
	v, err := t.Val(path)
	if sval, ok := v.(string); ok {
		return sval, err
	}
	return "", fmt.Errorf("StrVal found unexpected value type %T for path '%s'", v, path)
}

func (t STree) StrValMust(path string) string {
	v, err := t.StrVal(path)
	if err != nil {
		panic(err)
	}
	return v
}

// IntVal returns the value stored in data at the path, converting it
// to an int64, and returning the zero value if the int is not found.
func (t STree) IntVal(path string) (int64, error) {
	v, err := t.Val(path)
	if ival, ok := v.(int64); ok {
		return int64(ival), err
	} else if ival, ok := v.(int); ok {
		return int64(ival), err
	} else if ival, ok := v.(float64); ok {
		return int64(ival), err
	}
	return 0, fmt.Errorf("IntVal found unexpected value type %T for path '%s'", v, path)
}

func (t STree) IntValMust(path string) int64 {
	v, err := t.IntVal(path)
	if err != nil {
		panic(err)
	}
	return v
}

// FloatVal returns the value stored in data at the path, converting it
// to an int64, and returning the zero value if the value cannot be converted.
func (t STree) FloatVal(path string) (float64, error) {
	v, err := t.Val(path)
	if fval, ok := v.(float64); ok {
		return fval, err
	}
	return 0, fmt.Errorf("FloatVal found unexpected value type %T for path '%s'", v, path)
}

func (t STree) FloatValMust(path string) float64 {
	v, err := t.FloatVal(path)
	if err != nil {
		panic(err)
	}
	return v
}

// BVal returns the value stored in data at the path, converting it
// to an bool, and returning the zero value if the bool is not found.
func (t STree) BoolVal(path string) (bool, error) {
	v, err := t.Val(path)
	if bval, ok := v.(bool); ok {
		return bval, err
	}
	return false, fmt.Errorf("BoolVal found unexpected value type %T for path '%s'", v, path)
}

func (t STree) BoolValMust(path string) bool {
	v, err := t.BoolVal(path)
	if err != nil {
		panic(err)
	}
	return v
}

// STreeVal returns the value stored in data at the path, converting it
// to an STree and returning nil if the operation fails.
func (t STree) STreeVal(path string) (STree, error) {
	v, err := t.Val(path)
	if sval, ok := v.(STree); ok {
		return sval, err
	}
	return nil, fmt.Errorf("STreeVal found unexpected value type %T for path '%s'", v, path)
}

func (t STree) STreeValMust(path string) STree {
	v, err := t.STreeVal(path)
	if err != nil {
		panic(err)
	}
	return v
}

// SliceVal returns the value stored in the STree at the path, converting
// it to a []interface{} and returning nil if the operation fails.
func (t STree) SliceVal(path string) ([]interface{}, error) {
	v, err := t.Val(path)
	if aval, ok := v.([]interface{}); ok {
		return aval, err
	}
	return nil, fmt.Errorf("SliceVal found unexpected value type %T for path '%s'", v, path)
}

func (t STree) SliceValMust(path string) []interface{} {
	v, err := t.SliceVal(path)
	if err != nil {
		panic(err)
	}
	return v
}

// ValueOf assumes that the specified value is convertable to map[interface{}]interface{}
// and otherwise upholds the invariants of an STree structure. It's value as an STree
// is returned.
func ValueOf(v interface{}) (STree, error) {
	if sval, ok := v.(STree); ok {
		return sval, nil
	} else {
		return nil, fmt.Errorf("ValueOf failed to convert input (type %T)", v)
	}
}

// convertKeys returns the input map re-typed with all keys as interface{}
// wherever possible. This method facilitates use of the *Val methods for
// Unmarshaled json structures.
func convertKeys(input map[string]interface{}) (STree, error) {

	result := STree{}
	for k, v := range input {

		var iKey interface{} = k
		iVal, err := convertVal(v)
		if err != nil {
			return nil, err
		}
		result[iKey] = iVal
	}

	return result, nil
}

func convertVal(v interface{}) (interface{}, error) {

	var result interface{}

	val := reflect.ValueOf(v)
	if isPrimitiveKind(val.Kind()) {
		result = v

	} else if vSlice, ok := v.([]interface{}); ok {
		sVal := []interface{}{}
		for _, s := range vSlice {
			sConv, err := convertVal(s)
			if err != nil {
				return nil, err
			}
			sVal = append(sVal, sConv)
		}
		result = interface{}(sVal)

	} else if vMap, ok := v.(map[string]interface{}); ok {
		mVal, err := convertKeys(vMap)
		if err != nil {
			return nil, fmt.Errorf("convertVal error converting val: %v", vMap, err)
		}
		result = interface{}(mVal)

	} else {
		return nil, fmt.Errorf("convertVal unexpected type case")
	}

	return result, nil
}

// unconvertKeys returns a nested map with the same structure as the STree,
// but with string-typed keys, for use in json.Marshall() and the like.
func (s STree) unconvertKeys() (map[string]interface{}, error) {

	result := make(map[string]interface{})

	for k, v := range s {

		var kStr string
		if kStrVal, ok := k.(string); !ok {
			return nil, fmt.Errorf("unconvertKeys failed to convert key: %v", k)
		} else {
			kStr = kStrVal
		}

		val := reflect.ValueOf(v)
		if isPrimitiveKind(val.Kind()) {
			result[kStr] = v
		} else if vSlice, ok := v.([]interface{}); ok {
			result[kStr] = make([]interface{}, len(vSlice))
			for vIdx, vSub := range vSlice {
				if sVal, ok := vSub.(STree); ok {
					cVal, err := sVal.unconvertKeys()
					if err != nil {
						return nil, fmt.Errorf("unconvertKeys error converting slice key %s index %d: %v", k, vIdx, err)
					}
					result[kStr].([]interface{})[vIdx] = interface{}(cVal)
				}
			}

		} else if sVal, ok := v.(STree); ok {
			cVal, err := sVal.unconvertKeys()
			if err != nil {
				return nil, fmt.Errorf("unconvertKeys error converting key %s: %v", k, err)
			}
			result[kStr] = interface{}(cVal)
		} else {
			return nil, fmt.Errorf("unconvertKeys unexpected type case")
		}
	}

	return result, nil
}

// parsePathComponent parses the input as a stree key with an optional subscript
// index, e.g. "streeKey" or "treeList[2]". The key component and subscript index
// is returned, with -1 denoting no subscript present.
func (s STree) parsePathComponent(c string) (string, int, error) {

	path_comps := keyRegexp.FindStringSubmatch(c)
	if path_comps == nil || len(path_comps) < 1 {
		return "", -1, fmt.Errorf("parsePathComponent failed to parse path component %s", c)
	}

	path := path_comps[1]
	idx := -1
	if len(path_comps[2]) > 0 {
		i, err := strconv.Atoi(path_comps[2])
		if err != nil || i < 0 {
			return "", -1, fmt.Errorf("parsePathComponent failed to parse slice index %s from %s", path_comps[2], c)
		}
		idx = i
	}

	return path, idx, nil
}
