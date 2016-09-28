package stree

import (
	"fmt"
	"reflect"
)

// isPrimitive returns true if the specified Kind represents a primitive
// type, false otherwise.
func isPrimitive(k reflect.Kind) bool {
	return isBool(k) ||
		isInt(k) ||
		isUint(k) ||
		isFloat(k) ||
		isComplex(k) ||
		isString(k)
}

func isBool(k reflect.Kind) bool {
	return k == reflect.Bool
}

func isInt(k reflect.Kind) bool {
	return (k == reflect.Int ||
		k == reflect.Int8 ||
		k == reflect.Int16 ||
		k == reflect.Int32 ||
		k == reflect.Int64)
}

func isUint(k reflect.Kind) bool {
	return (k == reflect.Uint ||
		k == reflect.Uint8 ||
		k == reflect.Uint16 ||
		k == reflect.Uint32 ||
		k == reflect.Uint64)
}

func isFloat(k reflect.Kind) bool {
	return (k == reflect.Uintptr ||
		k == reflect.Float32 ||
		k == reflect.Float64)
}

func isComplex(k reflect.Kind) bool {
	return (k == reflect.Complex64 ||
		k == reflect.Complex128)
}

func isString(k reflect.Kind) bool {
	return k == reflect.String
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
