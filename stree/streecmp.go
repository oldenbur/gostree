package stree

import (
	"reflect"
	"strings"
)

type FieldComparisonResult int

const (
	COMP_NO_DIFFERENCE FieldComparisonResult = iota // the field exists and has the same value for both subject and object
	COMP_SUBJECT_LACKS                       = iota // the subject of the comparison lacks the field
	COMP_OBJECT_LACKS                        = iota // the object of the comparison lacks the field
	COMP_TYPES_DIFFER                        = iota // the type of the field differs between subject and object
	COMP_VALUES_DIFFER                       = iota // the value of the field differs between subject and object
)

func (f FieldComparisonResult) String() string {
	switch f {
	case COMP_NO_DIFFERENCE:
		return "COMP_NO_DIFFERENCE"
	case COMP_SUBJECT_LACKS:
		return "COMP_SUBJECT_LACKS"
	case COMP_OBJECT_LACKS:
		return "COMP_OBJECT_LACKS"
	case COMP_TYPES_DIFFER:
		return "COMP_TYPES_DIFFER"
	case COMP_VALUES_DIFFER:
		return "COMP_VALUES_DIFFER"
	default:
		return "UNKNOWN"
	}
}

type ComparisonResult map[string]FieldComparisonResult

func (s STree) CompareTo(o STree) ComparisonResult {

	result := map[string]FieldComparisonResult{}

	for _, f := range s.FieldPaths() {

		fStr := f.String()
		valSubj := s.Val(fStr)
		valObj := o.Val(fStr)

		if valObj == nil {
			result[fStr] = COMP_OBJECT_LACKS
			continue
		}

		kindSubj := reflect.ValueOf(valSubj).Kind()
		kindObj := reflect.ValueOf(valObj).Kind()

		if kindSubj != kindObj {
			result[fStr] = COMP_TYPES_DIFFER
		} else if isBool(kindSubj) {
			result[fStr] = compResult(s.BoolVal(fStr) == o.BoolVal(fStr))
		} else if isInt(kindSubj) || isUint(kindSubj) {
			result[fStr] = compResult(s.IntVal(fStr) == o.IntVal(fStr))
		} else if isFloat(kindSubj) {
			result[fStr] = compResult(s.FloatVal(fStr) == o.FloatVal(fStr))
		} else if isString(kindSubj) {
			result[fStr] = compResult(strings.Compare(s.StrVal(fStr), o.StrVal(fStr)) == 0)
		}

	}

	for _, f := range o.FieldPaths() {
		fStr := f.String()
		if s.Val(fStr) == nil {
			result[fStr] = COMP_SUBJECT_LACKS
		}
	}

	return result
}

func compResult(cond bool) (r FieldComparisonResult) {
	if cond {
		r = COMP_NO_DIFFERENCE
	} else {
		r = COMP_VALUES_DIFFER
	}
	return
}
