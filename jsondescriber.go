package jsondescriber

import (
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
func (o *RawObject) Inventory() (map[string]string, error) {
	var (
		inv  = make(map[string]string)
		obj  = *o
		typ  *string
		err  error
		lerr error
	)

	for k := range obj {
		if typ, err = TypeOf(obj[k]); err != nil {
			lerr = err
		}
		inv[k] = *typ
	}

	return inv, lerr
}

// Generates a friendly string from a JsonDescription, with Oxford comma and correct plurals
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
