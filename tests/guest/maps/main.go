package main

import "tests/maps/foo/foo/maps"

func init() {
	maps.Exports.NamedRoundtrip = maps.NamedRoundtrip
	maps.Exports.SameRoundtrip = maps.SameRoundtrip
	maps.Exports.AnonymousParam = maps.AnonymousParam
	maps.Exports.AnonymousRet = maps.AnonymousRet
	maps.Exports.Nested = maps.Nested
	maps.Exports.OptionMap = maps.OptionMap
	maps.Exports.ResultMap = maps.ResultMap
	maps.Exports.RecordMap = maps.RecordMap
	maps.Exports.VariantMap = maps.VariantMap
}

func main() {}
