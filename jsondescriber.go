package jsondescriber

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// Stores the type of an element and counts of its member element types, if applicable
type JsonDescription struct {
	Element string
	Members map[string]uint
}

// Constructor for JsonDescription that initializes its Members counter
func NewJsonDescription() *JsonDescription {
	return &JsonDescription{
		Element: "undefined",
		Members: make(map[string]uint),
	}
}

// A container for json.RawMessage from an array
type RawArray []json.RawMessage

// A container for json.RawMessage from an object
type RawObject map[string]json.RawMessage

// Each JSON element type except number is uniquely identifiable from its first character
var heuristics = map[string]string{
	`{`: `object`,
	`[`: `array`,
	`"`: `string`,
	`t`: `true`,
	`f`: `false`,
	`n`: `null`,
}

// Inverts a JsonDescription.Members into []"%uint %type(s)" with correct plurals
func descElem(counts map[string]uint) []string {
	var list = make([]string, 0)

	for k := range counts {
		count := counts[k]
		if count > 1 {
			desc := fmt.Sprintf("%d %ss", count, k)
			list = append(list, desc)
		} else if count == 1 {
			desc := fmt.Sprintf("%d %s", count, k)
			list = append(list, desc)
		}
	}

	return list
}

// Creates a key:type mapping from a RawObject for comparison
func (o *RawObject) Inventory() map[string]string {
	var (
		inv = make(map[string]string)
		obj = *o
		typ *string
	)

	for k := range obj {
		typ, _ = TypeOf(obj[k])
		inv[k] = *typ
	}

	return inv
}

// Generates a grammatical English-language list from a JsonDescription
func (jd *JsonDescription) Friendly() string {
	var descr string = "undefined"

	elem := jd.Element

	// Descriptions, not values
	if elem == "string" || elem == "number" {
		descr = fmt.Sprintf("a %s", elem)
	} else

	// Not to be confused with the string representation of that value
	if elem == "true" || elem == "false" || elem == "null" {
		descr = fmt.Sprintf("a literal %s", elem)
	} else

	// Type of container and inventory of elements; not concerned with keys here
	if elem == "object" || elem == "array" {
		inv := descElem(jd.Members)
		count := len(inv)

		// Oxfordize
		if count == 1 {
			descr = fmt.Sprintf(
				"an %s with %s",
				elem,
				inv[0],
			)
		} else if count == 2 {
			descr = fmt.Sprintf(
				"an %s with %s",
				elem,
				strings.Join(inv, " and "),
			)
		} else if count > 2 {
			descr = fmt.Sprintf(
				"an %s with %s, and %s",
				elem,
				strings.Join(inv[:count-1], ", "),
				inv[count-1],
			)
		} else {
			descr = fmt.Sprintf(
				"an empty %s",
				elem,
			)
		}
	}

	return descr
}

// Generates a populated JsonDescription from a raw JSON []byte
func Describe(data []byte) (*JsonDescription, error) {
	var (
		descr = NewJsonDescription()
		jt    *string
		err   error
	)

	// Bail if this isn't even JSON
	if jt, err = TypeOf(data); err != nil {
		return descr, err
	}

	descr.Element = *jt

	if descr.Element == "object" {
		jo := make(map[string]json.RawMessage)
		json.Unmarshal(data, &jo)

		for k := range jo {
			et, _ := TypeOf(jo[k])
			descr.Members[*et] += 1
		}
	}

	if descr.Element == "array" {
		ja := make(RawArray, 0)
		json.Unmarshal(data, &ja)

		for i := range ja {
			et, _ := TypeOf(ja[i])
			descr.Members[*et] += 1
		}
	}

	return descr, err
}

// Validates raw []byte as JSON and determines which element type it is
func TypeOf(data []byte) (*string, error) {
	var (
		typ string
		err error
	)

	if !json.Valid(data) {
		err = fmt.Errorf("not valid json")
		return &typ, err
	}

	if typ = heuristics[string(data[0])]; typ == "" {
		typ = "number"
	}

	return &typ, err
}

// this.Diff(that) maps keys of elements changed from this *RawObject to that one into four categories: added, deleted, modified, or typechanged
func (o *RawObject) Diff(n *RawObject) map[string][]string {
	var (
		add = make([]string, 0)
		del = make([]string, 0)
		mod = make([]string, 0)
		typ = make([]string, 0)
	)

	this := *o
	that := *n

	for k := range this {
		if that[k] == nil {
			del = append(del, k)
		} else {
			ot, _ := TypeOf(this[k])
			nt, _ := TypeOf(that[k])

			if *ot != *nt {
				typ = append(typ, k)
			} else if !bytes.Equal(this[k], that[k]) {
				mod = append(mod, k)
			}
		}
	}

	for k := range that {
		if this[k] == nil {
			add = append(add, k)
		}
	}

	return map[string][]string{
		"added":       add,
		"deleted":     del,
		"modified":    mod,
		"typechanged": typ,
	}
}

// this.DiffCount(that) counts members changed from this *RawObject to that one: added, deleted, modified, or typechanged
func (o *RawObject) DiffCount(n *RawObject) map[string]uint {
	diff := make(map[string]uint)

	this := *o
	that := *n

	for k := range this {
		if that[k] == nil {
			diff["deleted"] += 1
		} else {
			ot, _ := TypeOf(this[k])
			nt, _ := TypeOf(that[k])

			if *ot != *nt {
				diff["typechanged"] += 1
			} else if !bytes.Equal(this[k], that[k]) {
				diff["modified"] += 1
			}
		}
	}

	for k := range that {
		if this[k] == nil {
			diff["added"] += 1
		}
	}

	return diff
}

// UnmarshalArray is a convenience function wrapping json.Unmarshal to a new RawArray
func UnmarshalArray(in []byte) (*RawArray, error) {
	var (
		arr = new(RawArray)
		typ *string
		err error
	)

	typ, err = TypeOf(in)

	if err != nil {
		return arr, err
	}

	if *typ != "array" {
		err = fmt.Errorf("given []byte is %s, expected array", *typ)
		return arr, err
	}

	err = json.Unmarshal(in, &arr)
	return arr, err
}

// UnmarshalObject is a convenience function wrapping json.Unmarshal to a new RawObject
func UnmarshalObject(in []byte) (*RawObject, error) {
	var (
		obj = new(RawObject)
		typ *string
		err error
	)

	typ, err = TypeOf(in)

	if err != nil {
		return obj, err
	}

	if *typ != "object" {
		err = fmt.Errorf("given []byte is %s, expected object", *typ)
		return obj, err
	}

	err = json.Unmarshal(in, &obj)
	return obj, err
}
