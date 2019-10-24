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
