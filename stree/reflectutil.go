package stree

import (
	"fmt"
	"reflect"
)

func IsPrimitive(i interface{}) bool {
	return isPrimitiveKind(reflect.ValueOf(i).Kind())
}

// isPrimitiveKind returns true if the specified Kind represents a primitive
// type, false otherwise.
func isPrimitiveKind(k reflect.Kind) bool {
	return isBoolKind(k) ||
		isIntKind(k) ||
		isUintKind(k) ||
		isFloatKind(k) ||
		isComplexKind(k) ||
		isStringKind(k)
}

func IsBool(i interface{}) bool {
	return isBoolKind(reflect.ValueOf(i).Kind())
}

func isBoolKind(k reflect.Kind) bool {
	return k == reflect.Bool
}

func IsInt(i interface{}) bool {
	return isIntKind(reflect.ValueOf(i).Kind())
}

func isIntKind(k reflect.Kind) bool {
	return (k == reflect.Int ||
		k == reflect.Int8 ||
		k == reflect.Int16 ||
		k == reflect.Int32 ||
		k == reflect.Int64)
}

func IsUint(i interface{}) bool {
	return isUintKind(reflect.ValueOf(i).Kind())
}

func isUintKind(k reflect.Kind) bool {
	return (k == reflect.Uint ||
		k == reflect.Uint8 ||
		k == reflect.Uint16 ||
		k == reflect.Uint32 ||
		k == reflect.Uint64)
}

func IsFloat(i interface{}) bool {
	return isFloatKind(reflect.ValueOf(i).Kind())
}

func isFloatKind(k reflect.Kind) bool {
	return (k == reflect.Uintptr ||
		k == reflect.Float32 ||
		k == reflect.Float64)
}

func IsComplex(i interface{}) bool {
	return isComplexKind(reflect.ValueOf(i).Kind())
}

func isComplexKind(k reflect.Kind) bool {
	return (k == reflect.Complex64 ||
		k == reflect.Complex128)
}

func IsString(i interface{}) bool {
	return isStringKind(reflect.ValueOf(i).Kind())
}

func isStringKind(k reflect.Kind) bool {
	return k == reflect.String
}

func IsMap(i interface{}) bool {
	return (reflect.ValueOf(i).Kind() == reflect.Map)
}

func IsSlice(i interface{}) bool {
	return (reflect.ValueOf(i).Kind() == reflect.Slice)
}

func PrintValue(i interface{}) string {
	if i == nil {
		return "nil"
	} else {
		return printValue(reflect.ValueOf(i))
	}
}

func printValue(v reflect.Value) string {

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
