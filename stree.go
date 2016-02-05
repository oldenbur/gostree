package gostree

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"

	yaml "gopkg.in/yaml.v2"
	log "github.com/cihub/seelog"
)

type settingsMap map[string]*reflect.Value

type STree map[interface{}]interface{}

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

func PrintVal(v *reflect.Value) string {

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

// Val returns the leaf value at the position specified by path,
// which is a slash delimited list of nested keys in data, e.g.
// level1/level2/key
func (t STree) Val(path string) interface{} {

	keys := strings.Split(path, "/")
	log.Debugf("Val(%s) - %T", path, t)

	if len(keys) < 1 {
		return nil
	} else if len(keys) == 1 {
		log.Debugf("Val(%s) - LastKey: %v", path, t[keys[0]])
		return t[keys[0]]
	} else if data, ok := t[keys[0]].(STree); ok {
		return data.Val(strings.Join(keys[1:], "/"))
	} else if data, ok := t[keys[0]].([]interface{}); ok {
		log.Debugf("Val(%s) - slice: %v", path, data)
		return data
	} else {
		return nil
	}
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


// ConvertKeys returns the input map re-typed with all keys as interface{}
// wherever possible. This method facilitates use of the *Val methods for
// Unmarshaled json structures.
func convertKeys(input map[string]interface{}) (STree, error) {

	result := STree{}
	for k, v := range input {

		var iKey interface{} = k
		val := reflect.ValueOf(v)
		if isPrimitive(val.Kind()) {
			result[iKey] = v

		} else if vSlice, ok := v.([]interface{}); ok {
			sVal := []interface{}{}
			for _, s := range vSlice {
				sConv, err := convertVal(s)
				if err != nil {
					return nil, err
				}
				sVal = append(sVal, sConv)
			}
			result[iKey] = interface{}(sVal)

		} else if vMap, ok := v.(map[string]interface{}); ok {
			mVal, err := convertKeys(vMap)
			if err != nil {
				return nil, err
			}
			result[iKey] = interface{}(mVal)

		} else {
			return nil, fmt.Errorf("convertKeys unexpected type case for key %v", k)
		}
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
