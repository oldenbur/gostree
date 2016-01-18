package gostree

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	log "github.com/cihub/seelog"
	//"gopkg.in/yaml.v2"
)

type SettingsMap map[string]*reflect.Value

func init() {

	testConfig := `
        <seelog type="sync" minlevel="debug">
            <outputs formatid="main"><console/></outputs>
            <formats><format id="main" format="%Date %Time [%LEVEL] %Msg%n"/></formats>
        </seelog>`

	logger, err := log.LoggerFromConfigAsBytes([]byte(testConfig))
	if err != nil {
		panic(err)
	}

	err = log.ReplaceLogger(logger)
	if err != nil {
		panic(err)
	}
}

//func readYamlConfig(file string) (conf *CONF.AnubisConfig, err error) {
//
//	// Read the yaml file into memory
//	data, err := ioutil.ReadFile(configFile)
//
//	// Parse the yaml into a map
//	m := make(map[interface{}]interface{})
//	err = yaml.Unmarshal([]byte(data), &m)
//	if err != nil {
//		return nil, errors.New(fmt.Sprint("error in readYamlConfig yaml.Unmarshal: ", err))
//	}
//
//	conf = &CONF.AnubisConfig{
//		AccumulatorConf: &CONF.AnubisConfig_AccumulatorConf{
//			SecondsToAccumulate: IVal(m, "accumulatorConf/secondsToAccumulate"),
//			EntriesToAccumulate: IVal(m, "accumulatorConf/entriesToAccumulate"),
//			MaxBatchSizeBytes:   IVal(m, "accumulatorConf/maxBatchSizeBytes"),
//		},
//		GigawattDBConfig: &CONF.AnubisConfig_GigawattDBConfig{
//			GigawattDbPath: SVal(m, "gigawattDBConfig/gigawattDbPath"),
//		},
//		RelayConfig: &CONF.AnubisConfig_RelayConfig{
//			InputQueue:     SVal(m, "relayConfig/inputQueue"),
//			InputQueueSize: IVal(m, "relayConfig/inputQueueSize"),
//			DbMaxRows:      IVal(m, "relayConfig/dbMaxRows"),
//			DbResumeRows:   IVal(m, "relayConfig/dbResumeRows"),
//			DbMaxBytes:     IVal(m, "relayConfig/dbMaxBytes"),
//			DbResumeBytes:  IVal(m, "relayConfig/dbResumeBytes"),
//			DbCheckSizeMs:  IVal(m, "relayConfig/dbCheckSizeMs"),
//		},
//		BiffConfig: &CONF.AnubisConfig_BiffConfig{
//			AckQueue:     SVal(m, "biffConfig/ackQueue"),
//			AckQueueSize: IVal(m, "biffConfig/ackQueueSize"),
//		},
//		MessageRetryConfig: &CONF.AnubisConfig_MessageRetryConfig{
//			RetryIntervalSecs: IVal(m, "messageRetryConfig/retryIntervalSecs"),
//			MinRetrySecs:      IVal(m, "messageRetryConfig/minRetrySecs"),
//			MaxRetrySecs:      IVal(m, "messageRetryConfig/maxRetrySecs"),
//		},
//	}
//
//	return conf, err
//}

func findStructElemsPath(pre string, s interface{}, valsIn SettingsMap) (vals SettingsMap, err error) {

	vals = valsIn

	v := reflect.ValueOf(s)
	if v.Kind() != reflect.Ptr && v.Kind() != reflect.Interface {
		return vals, errors.New(fmt.Sprintf("findStructElems requires Ptr or Interface, got %s", v.Kind()))
	}

	r := v.Elem()
	log.Debugf("r.Type() = %s", r.Type())
	rType := r.Type()
	for i := 0; i < r.NumField(); i++ {

		f := r.Field(i)
		log.Debugf("%d: %s %s = %v", i, rType.Field(i).Name, f.Type(), f.Interface())

		if isPrimitive(f.Kind()) {
			vals[rType.Field(i).Name] = &f
		}

	}

	log.Debugf("vals: %+v", vals)
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
// which is a / delimited list of nested keys in data.
func Val(data map[interface{}]interface{}, path string) interface{} {

	keys := strings.Split(path, "/")

	if len(keys) < 1 {
		log.Warnf("Val ran out of keys")
		return nil
	} else if len(keys) == 1 {
		return data[keys[0]]
	} else if data, ok := data[keys[0]].(map[interface{}]interface{}); ok {
		return Val(data, strings.Join(keys[1:], "/"))
	} else {
		log.Warnf("Val failed on key %s", keys[0])
		return nil
	}
}

var BAD_STRING string = "##ERROR##"
var BAD_INT32 int32 = -9999

// SVal returns the value stored in data at the path, converting it
// to a pointer to string, and returning nil if the operation fails.
func SVal(data map[interface{}]interface{}, path string) *string {
	v := Val(data, path)
	if s, ok := v.(string); ok {
		return &s
	} else {
		log.Warnf("SVal failed to convert val %+v (%T) for key %s", v, v, path)
		return &BAD_STRING
	}
}

// IVal returns the value stored in data at the path, converting it
// to a pointer to int32, and returning nil if the operation fails.
func IVal(data map[interface{}]interface{}, path string) *int32 {
	v := Val(data, path)
	if i, ok := v.(int); ok {
		i32 := int32(i)
		return &i32
	} else {
		log.Warnf("IVal failed to convert val %+v (%T) for key %s", v, v, path)
		return &BAD_INT32
	}
}

// MVal returns the value stored in data at the path, converting it
// to a map and returning nil if the operation fails.
func MVal(data map[interface{}]interface{}, path string) map[interface{}]interface{} {
	v := Val(data, path)
	if m, ok := v.(map[interface{}]interface{}); ok {
		return m
	} else {
		log.Warnf("MVal failed to convert val %+v (%T) for key %s", v, v, path)
		return nil
	}
}

// ConvertKeys returns the input map re-typed with all keys as interface{}
// wherever possible. This method facilitates use of the *Val methods for
// Unmarshaled json structures.
func ConvertKeys(input map[string]interface{}) map[interface{}]interface{} {

	result := make(map[interface{}]interface{})

	for k, v := range input {

		var iKey interface{} = k
		val := reflect.ValueOf(v)
		if isPrimitive(val.Kind()) {
			result[iKey] = v
		} else if /*vSlice*/ _, ok := v.([]interface{}); ok {
			// leave array items out for now
		} else if vMap, ok := v.(map[string]interface{}); ok {
			result[iKey] = interface{}(ConvertKeys(vMap))
		} else {
			panic(fmt.Errorf("convertKeys unexpected type case"))
		}
	}

	return result
}
