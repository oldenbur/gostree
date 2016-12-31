package gostree

import (
	"testing"

	log "github.com/cihub/seelog"
	. "github.com/smartystreets/goconvey/convey"
)

func init() { configureTestLogger() }

func TestStruct(t *testing.T) {

	defer log.Flush()

	Convey("findStructElems", t, func() {

	})

	//	var yamlData string = `
	//---
	//product:
	//- sku         : BL394D
	//  quantity    : 4
	//  description : Basketball
	//  price       : 450.00
	//- sku         : BL4438H
	//  quantity    : 1
	//  description : Super Hoop
	//  price       : 2392.00
	//tax  : 251.42
	//total: 4443.52
	//comments: >
	//  Late afternoon is best.
	//  Backup contact is Nancy
	//  Billsmer @ 338-4338.
	//`
}
