package main

import (
	"fmt"
	"strings"
	"strconv"
	"regexp"
	"os"
	"flag"
	"bufio"
	
	"encoding/json"
	"github.com/clbanning/mxj"
	"github.com/naoina/toml"
	"gopkg.in/yaml.v2"
)

type complexStructure interface{}
type complexPath []interface{}

// add a single kv pair to a complex data structure
func addToComplex(s *complexStructure, path complexPath, value interface{}) {
	if len(path) == 0 {
		// base case
		*s = value
	} else {
		switch pathComponent := path[0].(type) {
			case nil:
				// array append case
				
				if *s == nil {
					*s = make([]complexStructure, 0, 1)
				}
				
				switch sSlice := (*s).(type) {
					case []complexStructure:
						var newElement complexStructure
						addToComplex(&newElement, path[1:], value)
						*s = append((*s).([]complexStructure), newElement)
					default:
						panic(fmt.Sprintf("s is %t, not slice\n", sSlice))
				}
			case int:
				// array with index case
				
				if *s == nil {
					*s = make([]complexStructure, 0, pathComponent+1)
				}
				
				switch sSlice := (*s).(type) {
					case []complexStructure:
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
					*s = make(map[string]complexStructure)
				}
				
				switch sMap := (*s).(type) {
					case map[string]complexStructure:
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
func splitPath(pathStr string) (path complexPath) {
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

func add(s *complexStructure, assignment string) {
	kv := strings.SplitN(assignment, "=", 2)
	path := splitPath(kv[0])
	value := guessType(kv[1], false)
	addToComplex(s, path, value)
}

func environ(pid int) (assignments []string) {
	if pid <= 0 {
		return os.Environ()
	}
	
	fileName := fmt.Sprintf("/proc/%d/environ", pid)
	file, err := os.Open(fileName)
	if err != nil {
		panic("Could not load environment for PID " + strconv.Itoa(pid) + "\n" +
		      "Failed to open " + fileName + ": " + err.Error())
	}
	
	reader := bufio.NewReader(file)
	// read null delineated file
	for {
		assignment, err := reader.ReadString('\x00')
		if err != nil {
			// break on EOF
			break
		}
		
		// remove null byte from end
		assignment = assignment[:len(assignment)-1]
		
		assignments = append(assignments, assignment)
	}
	return
}

func encode(obj interface{}, format string, xmlRootTag string) (string) {
	switch(format) {
		case "json":
			bytes, err := json.MarshalIndent(obj, "", "\t")
			if err != nil {
				panic(err)
			}
			return string(bytes)
		case "xml":
			bytes, err := mxj.AnyXmlIndent(obj, "", "\t", xmlRootTag)
			if err != nil {
				panic(err)
			}
			return string(bytes)
		case "toml":
			bytes, err := toml.Marshal(obj)
			if err != nil {
				panic(err)
			}
			return string(bytes)
		case "yaml":
			bytes, err := yaml.Marshal(obj)
			if err != nil {
				panic(err)
			}
			return string(bytes)
		default:
			panic("Format " + format + " is unsupported")
	}
}

func main() {
	// parse command line options
	var useUnderscores bool;
	flag.BoolVar(&useUnderscores, "underscores", false, "Use _ as a field seperator. Use __ (two underscores) for a literal _")
	var pid int;
	flag.IntVar(&pid, "pid", 0, "Read environment from given PID (defaults to self). This can be usefull if your shell strips environment variables containing special charicters.")
	var outputFormat string;
	flag.StringVar(&outputFormat, "fmt", "json", "Output format. Can be json, xml, toml, yaml")
	var xmlRootTag string;
	flag.StringVar(&xmlRootTag, "root", "config", "Root tag to be used when generating XML")
	flag.Parse()
	outputFormat = strings.ToLower(outputFormat);
	
	var s complexStructure
	
	var assignments []string
	if flag.NArg() == 1 {
		for _, assignment := range environ(pid) {
			prefix := flag.Arg(0)
			if strings.HasPrefix(assignment, prefix) {
				assignments = append(assignments, assignment[len(prefix):])
			}
		}
	} else {
		// no prefix
		assignments = environ(pid)
	}
	
	for _, assignment := range assignments {
		// exclude _= from the special underscores processing, as it is common and breaks things
		if useUnderscores && !strings.HasPrefix(assignment, "_=") {
			kv := strings.SplitN(assignment, "=", 2)
			// BUG(jon): there are no rules about how to parse 3 consecutive underscores
			// BUG(jon): Ideally we would not recognise the dot when in --underscores mode
			key := strings.Replace(kv[0], "_", ".", -1)
			key = strings.Replace(key, "..", "_", -1)
			assignment = key + "=" + kv[1]
		}
		add(&s, assignment)
	}
	
	// encode and print
	output := encode(s, outputFormat, xmlRootTag)
	fmt.Printf("%s\n", output)
}
