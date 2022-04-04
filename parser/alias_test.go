package parser

import (
	"reflect"
	"testing"
)

func TestGetNewAlias(t *testing.T) {
	result := &Alias{
		Name:        "__builtin__.int",
		PackageName: "testpackage",
		AliasOf:     "test",
	}
	alias := getNewAlias("int", "testpackage", "test")
	if !reflect.DeepEqual(alias, result) {
		t.Errorf("TestGetNewAlias: expected name to be %v got %v", result, alias)
	}
	result = &Alias{
		Name:        "TestStruct",
		PackageName: "testpackage",
		AliasOf:     "test",
	}
	alias = getNewAlias("TestStruct", "testpackage", "test")
	if !reflect.DeepEqual(alias, result) {
		t.Errorf("TestGetNewAlias: expected name to be %v got %v", result, alias)
	}
}

func TestAliasSlice(t *testing.T) {
	aliasSlice := AliasSlice{}
	aliasSlice = append(aliasSlice, Alias{
		Name:        "A",
		PackageName: "A",
		AliasOf:     "B",
	})
	aliasSlice = append(aliasSlice, Alias{
		Name:        "A",
		PackageName: "A",
		AliasOf:     "A",
	})
	if aliasSlice.Len() != 2 {
		t.Errorf("TestAliasSlice: Expected len of slice = 2, got %d", aliasSlice.Len())
	}
	if aliasSlice.Less(0, 1) {
		t.Errorf("TestAliasSlice: Expected Less(0,1) to be false, got %t", aliasSlice.Less(0, 1))
	}
	aliasSlice.Swap(0, 1)
	if aliasSlice[0].AliasOf != "A" {
		t.Errorf("TestAliasSlice: Expected aliasSlice[0].AliasOf to be 'A' got %s", aliasSlice[0])
	}
}
