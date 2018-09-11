package main

import (
	"fmt"
	"encoding/json"
	"strings"
	"strconv"
	"regexp"
	"os"
)

type jsonStructure interface{}
type jsonPath []interface{}

// add a single kv pair to a complex data structure
func addToComplex(s *jsonStructure, path jsonPath, value interface{}) {
	if len(path) == 0 {
		// base case
		*s = value
	} else {
		switch pathComponent := path[0].(type) {
			case nil:
				// array append case
				
				if *s == nil {
					*s = make([]jsonStructure, 0, 1)
				}
				
				switch sSlice := (*s).(type) {
					case []jsonStructure:
						var newElement jsonStructure
						addToComplex(&newElement, path[1:], value)
						*s = append((*s).([]jsonStructure), newElement)
					default:
						panic(fmt.Sprintf("s is %t, not slice\n", sSlice))
				}
			case int:
				// array with index case
				
				if *s == nil {
					*s = make([]jsonStructure, 0, pathComponent+1)
				}
				
				switch sSlice := (*s).(type) {
					case []jsonStructure:
						for len(sSlice) <= pathComponent {
							sSlice = append(sSlice, nil)
							*s = sSlice
						}
						workingCopy := sSlice[pathComponent]
						addToComplex(&workingCopy, path[1:], value)
						sSlice[pathComponent] = workingCopy
					default:
						panic(fmt.Sprintf("s is %t, not slice\n", sSlice))
				}
			case string:
				// map case
				
				if *s == nil {
					*s = make(map[string]jsonStructure)
				}
				
				switch sMap := (*s).(type) {
					case map[string]jsonStructure:
						if elem, exists := sMap[pathComponent]; !exists {
							sMap[pathComponent] = elem
						}
						workingCopy := sMap[pathComponent]
						addToComplex(&workingCopy, path[1:], value)
						sMap[pathComponent] = workingCopy
					default:
						panic(fmt.Sprintf("s is %t, not map\n", sMap))
				}
			default:
				panic(fmt.Sprintf("pathComponent is of unknown type: %t\n", pathComponent))
		}
	}
}

func guessType(val string, emptyNil bool) (ret interface{}) {
	// can't use json because of case
	if strings.ToLower(val) == "true" {
		return true
	}
	
	// can't use json because of case
	if strings.ToLower(val) == "false" {
		return false
	}
	
	if strings.ToLower(val) == "nil" {
		return nil
	}
	
	// can't use json because of case
	if strings.ToLower(val) == "null" {
		return nil
	}
	
	if val == "" && emptyNil {
		return nil
	}
	
	// try to get a 32 bit int, fall back to json's 64 bit if we can't
	if i, err := strconv.ParseInt(val, 10, 32); err == nil {
		return int(i)
	}
	
	if err := json.Unmarshal([]byte(val), &ret); err == nil {
		return
	}
	
	// plain string
	return val
}

// BUG(jon): splitPath() can break silently if an identifier contains special chars
// ex: foo.bar["weird[identifier]"].baz
func splitPath(pathStr string) (path []interface{}) {
	tokenRegex := regexp.MustCompile(`(?:^|\.)(?:([^[\].]+)((?:\[[^[\].]*\])*)|)`)
	indexRegex := regexp.MustCompile(`\[([^[\]]*)\]`)
	matches := tokenRegex.FindAllStringSubmatch(pathStr, -1)
	for _, match := range matches {
		// the normal identifier part
		token := guessType(match[1], true)
		path = append(path, token)
		
		// the brackets part
		indexes := indexRegex.FindAllStringSubmatch(match[2], -1)
		for _, index := range indexes {
			token := guessType(index[1], true)
			path = append(path, token)
		}
	}
	return
}

// BUG(jon): add() can break silently if the path contains an '='.
// this bug does not affect the value.
// broken: foo.bar["="]="baz"
// ok:     foo.bar="fiz=buz"
func add(s *jsonStructure, assignment string) {
	kv := strings.SplitN(assignment, "=", 2)
	path := splitPath(kv[0])
	value := guessType(kv[1], false)
	addToComplex(s, path, value)
}

func main() {
	var s jsonStructure
	
	var assignments []string
	if len(os.Args) == 2 {
		for _, assignment := range os.Environ() {
			prefix := os.Args[1]
			if len(assignment) > len(prefix) && assignment[:len(prefix)] == prefix {
				assignments = append(assignments, assignment[len(prefix):])
			}
		}
	} else {
		// no prefix
		assignments = os.Environ()
	}
	
	for _, assignment := range assignments {
		add(&s, assignment)
	}
	
	b, _ := json.MarshalIndent(s, "", "\t")
	fmt.Printf("%s\n", b)
}
