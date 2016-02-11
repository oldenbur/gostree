package gostree

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	yaml "gopkg.in/yaml.v2"
	log "github.com/cihub/seelog"
)

type settingsMap map[string]*reflect.Value

type STree map[interface{}]interface{}

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

// isPrimitive returns true if the specified Kind represents a primitive
// type, false otherwise.
func isPrimitive(k reflect.Kind) bool {
	return (k == reflect.Bool ||
		k == reflect.Int ||
		k == reflect.Int8 ||
		k == reflect.Int16 ||
		k == reflect.Int32 ||
		k == reflect.Int64 ||
		k == reflect.Uint ||
		k == reflect.Uint8 ||
		k == reflect.Uint16 ||
		k == reflect.Uint32 ||
		k == reflect.Uint64 ||
		k == reflect.Uintptr ||
		k == reflect.Float32 ||
		k == reflect.Float64 ||
		k == reflect.Complex64 ||
		k == reflect.Complex128 ||
		k == reflect.String)
}

func printVal(v reflect.Value) string {

	switch v.Kind() {
	case reflect.Bool:
		return fmt.Sprintf("%t", v.Bool())
	case reflect.String:
		return v.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%d", v.Uint())
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%f", v.Float())
	case reflect.Complex64, reflect.Complex128:
		return fmt.Sprintf("%+v", v.Complex())
	default:
		return fmt.Sprintf("<printVal: %v>", v)
	}
}

// keyRegexp matches strings of the form key_name or slice_name[123]
var keyRegexp *regexp.Regexp = regexp.MustCompile(`^(\w+)(?:\[(\d+)\])?$`)

// Val returns the leaf value at the position specified by path,
// which is a slash delimited list of nested keys in data, e.g.
// level1/level2/key
func (t STree) Val(path string) interface{} {

	keys := strings.Split(path, "/")
//	log.Debugf("Val(%s) - %T", path, t)

	key_comps := keyRegexp.FindStringSubmatch(keys[0])
	if key_comps == nil || len(key_comps) < 1 {
		log.Warnf("val failed to parse key %s", keys[0])
		return nil
	}

	key := key_comps[1]
	idx := -1
	if len(key_comps[2]) > 0 {
		i, err := strconv.Atoi(key_comps[2])
		if err != nil || i < 0 {
			log.Warnf("val failed to parse slice index %s", key_comps[1])
			return nil
		}
		idx = i
	}

	if len(keys) < 1 {
		return nil

	} else if len(keys) == 1 && idx < 0 {
//		log.Debugf("Val(%s) - LastKey: %v", path, t[key])
		return t[key]

	} else if data, ok := t[key].(STree); ok {
		if idx >= 0 {
			log.Warnf("Val unexpected index for STree value: %s", keys[0])
			return nil
		}
		return data.Val(strings.Join(keys[1:], "/"))

	} else if data, ok := t[key].([]interface{}); ok {
		// TODO: break this case out to recursively handle nested slices
//		log.Debugf("Val(%s) - slice: %v", path, data)
		if idx >= 0 && idx < len(data) {
			result := data[idx]
			if len(keys) < 2 {
				return result
			} else if sval, ok := result.(STree); ok {
				return sval.Val(strings.Join(keys[1:], "/"))
			}

		} else if idx < 0 {
			if len(keys) > 1 {
				log.Warnf("Val requires index to traverse slice value for key: %s", keys[0])
				return nil
			}
			return data

		} else {
			log.Warnf("Val invalid slice key index: %s", keys[0])
			return nil
		}
	}

	log.Warnf("Val failed to produce value for key: %s", keys[0])
	return nil
}

// SVal returns the value stored in data at the path, converting it
// to a a string, and returning the zero value if the string is not
// found.
func (t STree) StrVal(path string) (s string) {
	v := t.Val(path)
	if sval, ok := v.(string); ok {
		s = sval
	}
	return
}

// IVal returns the value stored in data at the path, converting it
// to an int64, and returning the zero value if the int is not found.
func (t STree) IntVal(path string) (i int64) {
	v := t.Val(path)
	if ival, ok := v.(int64); ok {
		i64 := int64(ival)
		i = i64
	} else if ival, ok := v.(float64); ok {
		i64 := int64(ival)
		i = i64
	}
	return
}

// BVal returns the value stored in data at the path, converting it
// to an bool, and returning the zero value if the bool is not found.
func (t STree) BoolVal(path string) (b bool) {
	v := t.Val(path)
	if bval, ok := v.(bool); ok {
		b = bval
	}
	return
}

// TVal returns the value stored in data at the path, converting it
// to an STree and returning nil if the operation fails.
func (t STree) STreeVal(path string) (s STree) {
	v := t.Val(path)
	if sval, ok := v.(STree); ok {
		s = sval
	}
	return
}


func (t STree) SliceVal(path string) (a []interface{}) {
	v := t.Val(path)
	if aval, ok := v.([]interface{}); ok {
		a = aval
	}
	return
}

func ValueOf(v interface{}) (STree, error) {
	if sval, ok := v.(STree); ok {
		return sval, nil
	} else {
		return nil, fmt.Errorf("ValueOf failed to convert input (type %T)", v)
	}
}

// MarhsalJSON returns a JSON-rendered representation of the subject STree.
func (t STree) MarshalJSON() ([]byte, error) {

	buf := []byte{'{'}
	i := 0

	for k, v := range t {

		if i > 0 {
			buf = append(buf, ',')
		}
		buf = append(buf, []byte(fmt.Sprintf(`"%v":`, k))...)

		vMarsh, err := marshalJSONVal(v)
		if err != nil {
			return nil, err
		}
		buf = append(buf, vMarsh...)

		i += 1
	}
	buf = append(buf, '}')

	return buf, nil
}

// marshalJSONVal examines the structure of the specified value and returns
// a JSON-rendered representation of it.
func marshalJSONVal(v interface{}) ([]byte, error) {

	val := reflect.ValueOf(v)
	if val.Kind() == reflect.String {
		sVal := strings.Replace(val.String(), `\`, `\\`, -1)
		return []byte(fmt.Sprintf(`"%s"`, sVal)), nil

	} else if isPrimitive(val.Kind()) {
		return []byte(fmt.Sprintf(`%s`, printVal(val))), nil

	} else if vSlice, ok := v.([]interface{}); ok {
		buf := []byte{}
		buf = append(buf, '[')

		for i, sv := range vSlice {
			m, err := marshalJSONVal(sv)
			if err != nil {
				return nil, err
			}
			if i > 0 {
				buf = append(buf, ',')
			}
			buf = append(buf, m...)
		}
		buf = append(buf, ']')
		return buf, nil

	} else if vSTree, ok := v.(STree); ok {
		buf, err := vSTree.MarshalJSON()
		if err != nil {
			return nil, err
		}
		return buf, nil
	}

	return nil, fmt.Errorf("marshalJSONVal unhandled value type for %v", v)
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


