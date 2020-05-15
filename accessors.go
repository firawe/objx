package objx

import (
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

const (
	// PathSeparator is the character used to separate the elements
	// of the keypath.
	//
	// For example, `location\address\city`
	PathSeparator string = "\\"

	// arrayAccesRegexString is the regex used to extract the array number
	// from the access path
	arrayAccesRegexString = `^(.+)\[([0-9]+)\]$`

	// mapAccessRegexString is the regex used to extract the map key
	// from the access path
	mapAccessRegexString = `^([^\[]*)\[([^\]]+)\](.*)$`
)

// arrayAccesRegex is the compiled arrayAccesRegexString
var arrayAccesRegex = regexp.MustCompile(arrayAccesRegexString)

// mapAccessRegex is the compiled mapAccessRegexString
var mapAccessRegex = regexp.MustCompile(mapAccessRegexString)

// Get gets the value using the specified selector and
// returns it inside a new Obj object.
//
// If it cannot find the value, Get will return a nil
// value inside an instance of Obj.
//
// Get can only operate directly on map[string]interface{} and []interface.
//
// Example
//
// To access the title of the third chapter of the second book, do:
//
//    o.Get("books[1].chapters[2].title")
func (m Map) Get(selector string, opts ...*Option) *Value {
	rawObj := access(m, selector, nil, false, opts...)
	return &Value{data: rawObj}
}

// Set sets the value using the specified selector and
// returns the object on which Set was called.
//
// Set can only operate directly on map[string]interface{} and []interface
//
// Example
//
// To set the title of the third chapter of the second book, do:
//
//    o.Set("books[1].chapters[2].title","Time to Go")
func (m Map) Set(selector string, value interface{}, opts ...*Option) Map {
	access(m, selector, value, true, opts...)
	return m
}

// getIndex returns the index, which is hold in s by two braches.
// It also returns s withour the index part, e.g. name[1] will return (1, name).
// If no index is found, -1 is returned
func getIndex(s string) (int, string) {
	arrayMatches := arrayAccesRegex.FindStringSubmatch(s)
	if len(arrayMatches) > 0 {
		// Get the key into the map
		selector := arrayMatches[1]
		// Get the index into the array at the key
		// We know this cannt fail because arrayMatches[2] is an int for sure
		index, _ := strconv.Atoi(arrayMatches[2])
		return index, selector
	}
	return -1, s
}

// getKey returns the key which is held in s by two brackets.
// It also returns the next selector.
func getKey(s string) (string, string) {
	selSegs := strings.SplitN(s, PathSeparator, 2)
	thisSel := selSegs[0]
	nextSel := ""

	if len(selSegs) > 1 {
		nextSel = selSegs[1]
	}

	mapMatches := mapAccessRegex.FindStringSubmatch(s)
	if len(mapMatches) > 0 {
		// check if match between [ and ] is not a number (then it would be an array access rather than map)
		if _, err := strconv.Atoi(mapMatches[2]); err != nil {
			thisSel = mapMatches[1]
			nextSel = "[" + mapMatches[2] + "]" + mapMatches[3]

			if thisSel == "" {
				thisSel = mapMatches[2]
				nextSel = mapMatches[3]
			}

			if nextSel == "" {
				selSegs = []string{"", ""}
			} else if string(nextSel[0]) == PathSeparator {
				nextSel = nextSel[1:]
			}
		}
	}
	return thisSel, nextSel
}

// access accesses the object using the selector and performs the
// appropriate action.
func access(current interface{}, selector string, value interface{}, isSet bool, opts ...*Option) interface{} {
	thisSel, nextSel := getKey(selector)

	index := -1
	if strings.Contains(thisSel, "[") {
		index, thisSel = getIndex(thisSel)
	}

	switch current.(type) {
	case Map:
		current = map[string]interface{}(current.(Map))
	case map[string]interface{}, []interface{}:
		break
	default:
		valueOfCurrent := reflect.ValueOf(current)
		if valueOfCurrent.IsValid() {
			for _, opt := range opts {
				if opt.Value.Type().Name() == valueOfCurrent.Type().Name() {
					current = opt.ConvertToMap(current)
				}
			}
		}
	}

	// get the object in question
	curMSI := map[string]interface{}{}
	switch current.(type) {
	case map[string]interface{}:
		curMSI = current.(map[string]interface{})
		if nextSel == "" && isSet && index == -1 {
			curMSI[thisSel] = value
			return nil
		}

		_, ok := curMSI[thisSel].(map[string]interface{})
		if !ok {
			// try to cast to objx.Map instead
			_, ok = curMSI[thisSel].(Map)
		}
		if (curMSI[thisSel] == nil || !ok) && index == -1 && isSet {
			curMSI[thisSel] = map[string]interface{}{}
		}

		current = curMSI[thisSel]
	default:
		current = nil
	}

	// do we need to access the item of an array?
	if index > -1 {
		if array, ok := interSlice(current); ok {
			if index < len(array) {
				if isSet && nextSel == "" {
					array[index] = value
				} else {
					current = array[index]
				}
			} else {
				if isSet && nextSel == "" {
					curMSI[thisSel] = append(array, value) // set appended array to selector position to keep pointer to original map
				} else {
					current = nil
				}
			}
		}
	}
	if nextSel != "" {
		current = access(current, nextSel, value, isSet, opts...)
	}
	return current
}

func interSlice(slice interface{}) ([]interface{}, bool) {
	if array, ok := slice.([]interface{}); ok {
		return array, ok
	}

	s := reflect.ValueOf(slice)
	if s.Kind() != reflect.Slice {
		return nil, false
	}

	ret := make([]interface{}, s.Len())

	for i := 0; i < s.Len(); i++ {
		ret[i] = s.Index(i).Interface()
	}

	return ret, true
}
