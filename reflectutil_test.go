package gostree

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func init() { configureTestLogger() }

func TestReflectUtil(t *testing.T) {

	Convey("IsPrimitive", t, func() {
		So(IsPrimitive(3), ShouldBeTrue)
		So(IsPrimitive("abc"), ShouldBeTrue)
		So(IsPrimitive(false), ShouldBeTrue)
		So(IsPrimitive(12.34), ShouldBeTrue)

		So(IsPrimitive([]int{1, 2}), ShouldBeFalse)
	})

	Convey("IsBool", t, func() {
		So(IsBool(true), ShouldBeTrue)
		So(IsBool(false), ShouldBeTrue)

		So(IsBool(1), ShouldBeFalse)
	})

	Convey("IsInt", t, func() {
		So(IsInt(-12), ShouldBeTrue)
		So(IsInt(48), ShouldBeTrue)

		So(IsInt("a"), ShouldBeFalse)
	})

	Convey("IsUint", t, func() {
		So(IsUint(uint(0)), ShouldBeTrue)
		So(IsUint(uint(48)), ShouldBeTrue)

		So(IsUint(-12), ShouldBeFalse)
		So(IsUint("a"), ShouldBeFalse)
	})

	Convey("IsFloat", t, func() {
		So(IsFloat(12.34), ShouldBeTrue)
		So(IsFloat(-14.), ShouldBeTrue)

		So(IsFloat(-12), ShouldBeFalse)
		So(IsFloat("a"), ShouldBeFalse)
	})

	Convey("IsString", t, func() {
		So(IsString("abc"), ShouldBeTrue)
		So(IsString(""), ShouldBeTrue)

		So(IsString(-12), ShouldBeFalse)
		So(IsString(true), ShouldBeFalse)
	})

	Convey("IsMap", t, func() {
		So(IsMap(map[string]int{"a": 1}), ShouldBeTrue)
		So(IsMap(map[string]interface{}{"a": 1}), ShouldBeTrue)
		So(IsMap(map[interface{}]interface{}{4: 1}), ShouldBeTrue)

		So(IsMap(-12), ShouldBeFalse)
		So(IsMap(true), ShouldBeFalse)
	})

	Convey("IsSlice", t, func() {
		So(IsSlice([]int{1, 2, 3}), ShouldBeTrue)
		So(IsSlice([]string{"a", "b"}), ShouldBeTrue)
		So(IsSlice([]interface{}{"a", 1}), ShouldBeTrue)

		So(IsSlice(-12), ShouldBeFalse)
		So(IsSlice(true), ShouldBeFalse)
	})

	Convey("PrintValue", t, func() {
		So(PrintValue(nil), ShouldEqual, "nil")
		So(PrintValue(true), ShouldEqual, "true")
		So(PrintValue("def"), ShouldEqual, "def")
		So(PrintValue(-678), ShouldEqual, "-678")
		So(PrintValue(uint16(987)), ShouldEqual, "987")
		So(PrintValue(-87.65), ShouldEqual, "-87.650000")
		So(PrintValue([]int{1, 2, 3}), ShouldEqual, "[1 2 3]")
	})

}
