package stree

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	//	log "github.com/cihub/seelog"
	yaml "gopkg.in/yaml.v2"
)

type settingsMap map[string]*reflect.Value

type STree map[interface{}]interface{}

type FieldPath []string

func (p FieldPath) String() string {
	return strings.Join([]string(p), "/")
}

func ValueOfPath(p string) FieldPath {
	return FieldPath(strings.Split(p, "/"))
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

		if isPrimitive(f.Kind()) {
			vals[rType.Field(i).Name] = &f
		}
	}

	return vals, nil
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
var keyRegexp *regexp.Regexp = regexp.MustCompile(`^(\w+)(?:\[(\d+)\])?$`)

// Val returns the leaf value at the position specified by path,
// which is a slash delimited list of nested keys in data, e.g.
// level1/level2/key
func (t STree) Val(path string) (interface{}, error) {

	keys := strings.Split(path, "/")
	//	log.Debugf("Val(%s) - %T", path, t)

	key_comps := keyRegexp.FindStringSubmatch(keys[0])
	if key_comps == nil || len(key_comps) < 1 {
		return nil, fmt.Errorf("val failed to parse key %s", keys[0])
	}

	key := key_comps[1]
	idx := -1
	if len(key_comps[2]) > 0 {
		i, err := strconv.Atoi(key_comps[2])
		if err != nil || i < 0 {
			return nil, fmt.Errorf("val failed to parse slice index %s", key_comps[1])
		}
		idx = i
	}

	if len(keys) < 1 {
		return nil, fmt.Errorf("no key remaining components")

	} else if len(keys) == 1 && idx < 0 {
		//		log.Debugf("Val(%s) - LastKey: %v", path, t[key])
		return t[key], nil

	} else if data, ok := t[key].(STree); ok {
		if idx >= 0 {
			return nil, fmt.Errorf("Val unexpected index for STree value: %s", keys[0])
		} else {
			return data.Val(strings.Join(keys[1:], "/"))
		}

	} else if data, ok := t[key].([]interface{}); ok {
		// TODO: break this case out to recursively handle nested slices
		//		log.Debugf("Val(%s) - slice: %v", path, data)
		if idx >= 0 && idx < len(data) {
			result := data[idx]
			if len(keys) < 2 {
				return result, nil
			} else if sval, ok := result.(STree); ok {
				return sval.Val(strings.Join(keys[1:], "/"))
			}

		} else if idx < 0 {
			if len(keys) > 1 {
				return nil, fmt.Errorf("Val requires index to traverse slice value for key: %s", keys[0])
			}
			return data, nil

		} else {
			return nil, fmt.Errorf("Val invalid slice key index: %s", keys[0])
		}
	}

	return nil, fmt.Errorf("Val failed to produce value for key: %s", keys[0])
}

// SVal returns the value stored in data at the path, converting it
// to a a string, and returning the zero value if the string is not
// found.
func (t STree) StrVal(path string) (string, error) {
	v, err := t.Val(path)
	if sval, ok := v.(string); ok {
		return sval, err
	}
	return "", err
}

// IntVal returns the value stored in data at the path, converting it
// to an int64, and returning the zero value if the int is not found.
func (t STree) IntVal(path string) (int64, error) {
	v, err := t.Val(path)
	if ival, ok := v.(int64); ok {
		return int64(ival), err
	} else if ival, ok := v.(float64); ok {
		return int64(ival), err
	}
	return 0, err
}

// FloatVal returns the value stored in data at the path, converting it
// to an int64, and returning the zero value if the value cannot be converted.
func (t STree) FloatVal(path string) (float64, error) {
	v, err := t.Val(path)
	if fval, ok := v.(float64); ok {
		return fval, err
	}
	return 0, err
}

// BVal returns the value stored in data at the path, converting it
// to an bool, and returning the zero value if the bool is not found.
func (t STree) BoolVal(path string) (bool, error) {
	v, err := t.Val(path)
	if bval, ok := v.(bool); ok {
		return bval, err
	}
	return false, err
}

// STreeVal returns the value stored in data at the path, converting it
// to an STree and returning nil if the operation fails.
func (t STree) STreeVal(path string) (STree, error) {
	v, err := t.Val(path)
	if sval, ok := v.(STree); ok {
		return sval, err
	}
	return nil, err
}

// SliceVal returns the value stored in the STree at the path, converting
// it to a []interface{} and returning nil if the operation fails.
func (t STree) SliceVal(path string) ([]interface{}, error) {
	v, err := t.Val(path)
	if aval, ok := v.([]interface{}); ok {
		return aval, err
	}
	return nil, err
}

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
	if isPrimitive(val.Kind()) {
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
		if isPrimitive(val.Kind()) {
			result[kStr] = v
		} else if /*vSlice*/ _, ok := v.([]interface{}); ok {
			// leave array items out for now
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
