package stree

import (
	"reflect"
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

var errTracker error

func (s STree) CompareTo(o STree) (ComparisonResult, error) {

	result := map[string]FieldComparisonResult{}

	for _, f := range s.FieldPaths() {

		fStr := f.String()
		valSubj, err := s.Val(fStr)
		if err != nil {
			return nil, err
		}

		valObj, err := o.Val(fStr)
		if valObj == nil {
			result[fStr] = COMP_OBJECT_LACKS
			continue
		}

		kindSubj := reflect.ValueOf(valSubj).Kind()
		kindObj := reflect.ValueOf(valObj).Kind()

		if kindSubj != kindObj {
			result[fStr] = COMP_TYPES_DIFFER
		} else if valObj == valSubj {
			result[fStr] = COMP_NO_DIFFERENCE
		} else {
			result[fStr] = COMP_VALUES_DIFFER
		}

	}

	for _, f := range o.FieldPaths() {
		fStr := f.String()
		if sVal, _ := s.Val(fStr); sVal == nil {
			result[fStr] = COMP_SUBJECT_LACKS
		}
	}

	return result, nil
}

func compResult(cond bool) (r FieldComparisonResult) {
	if cond {
		r = COMP_NO_DIFFERENCE
	} else {
		r = COMP_VALUES_DIFFER
	}
	return
}
