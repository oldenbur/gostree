package stree

import (
	"os"
	"path"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTestUtil(t *testing.T) {

	Convey("Test RandomString", t, func() {

		strLen := 64
		s1 := RandomString(strLen)
		s2 := RandomString(strLen)

		So(len(s1), ShouldEqual, strLen)
		So(len(s2), ShouldEqual, strLen)
	})

	Convey("Test PackageDirectory", t, func() {

		dir, err := PackageDirectory()
		So(err, ShouldBeNil)

		_, err = os.Stat(path.Join(dir, "strings_test.go"))
		So(os.IsNotExist(err), ShouldBeFalse)
	})
}
