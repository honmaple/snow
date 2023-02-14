package lowervar

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
)

func TestToLower(t *testing.T) {
	type inlineMap map[string]string
	type inlineStruct struct {
		Val  int      `json:"val"`
		Val1 []string `json:"val1"`
	}
	type testStruct struct {
		Val   string                 `json:"val"`
		Slice []*inlineStruct        `json:"slice"`
		Map   map[string]interface{} `json:"map"`
		inlineStruct
		inlineMap
	}

	testFunc := func() string {
		return "Hello"
	}
	data := [][2]interface{}{
		[2]interface{}{
			&inlineStruct{10, []string{"2", "1"}},
			map[string]interface{}{
				"val":  10,
				"val1": []interface{}{"2", "1"},
			},
		},
		[2]interface{}{
			&testStruct{Val: "1", Slice: []*inlineStruct{&inlineStruct{Val: 20}, &inlineStruct{Val: 50}}},
			map[string]interface{}{
				"val":   0,
				"slice": []interface{}{map[string]interface{}{"val": 20, "val1": nil}, map[string]interface{}{"val": 50, "val1": nil}},
				"map":   nil,
				"val1":  nil,
			},
		},
		[2]interface{}{
			&testStruct{Val: "1", Map: map[string]interface{}{
				"val":  "1",
				"val1": &inlineStruct{Val: 80},
				"val3": testFunc,
			}},
			map[string]interface{}{
				"val":   "1",
				"map":   map[string]interface{}{"val": 1, "val1": map[string]interface{}{"val": 80, "val1": nil}, "val3": testFunc},
				"slice": nil,
				"val1":  "",
			},
		},
		// [2]interface{}{
		//	&testStruct{
		//		inlineStruct: inlineStruct{Val1: []string{"3", "1"}},
		//		inlineMap:    map[string]string{"val2": "aaa"},
		//	},
		//	map[string]interface{}{
		//		"val":   0,
		//		"slice": nil,
		//		"map":   nil,
		//		"val1":  []interface{}{"3", "1"},
		//		// "val2":  "aaa",
		//	},
		// },
	}

	for _, d := range data {
		var m map[string]interface{}
		dd, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{TagName: "json", Result: &m})
		dd.Decode(d[0])
		fmt.Println(m)
		assert.Equal(t, d[1], toLower(reflect.ValueOf(d[0]), "json"))
	}
}
