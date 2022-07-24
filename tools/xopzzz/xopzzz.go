package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"
)

var macroRE = regexp.MustCompile(`^(\s*)//\s?MACRO (\w+)(?:\s+SKIP:(\S+))?\s*$`)
var errorRE = regexp.MustCompile(`^(\s*)//MACRO/`)
var indentRE = regexp.MustCompile(`^(\s*)(?:\S|$)`)
var zzzRE = regexp.MustCompile(`(zzz|ZZZ)`)
var packageRE = regexp.MustCompile(`^package (\w+)`)

var macros = map[string]map[string]string{
	// ZZZAttribute are the span-attributes in the main logger.  These
	// turn into Span methods like Span.BoolAttribute() and a corresponding
	// xopconst.BoolAttribute type.
	"ZZZAttribute": {
		"Bool":     "bool",
		"Number":   "float64", // XXX add
		"Float32":  "float32", // XXX add
		"Int64":    "int64",
		"Int32":    "int32",
		"Int16":    "int16",
		"Int8":     "int8",
		"Int":      "int",
		"Str":      "string",
		"Link":     "trace.Trace",
		"Any":      "interface{}",
		"Time":     "time.Time",
		"Duration": "time.Duration",
		"Enum":     "xopconst.Enum",
	},
	// BaseAttributes are the span decorators that base loggers must
	// implement.  These turn into things like Base.MetadataBool()
	"BaseAttribute": {
		"Bool":   "bool",
		"Number": "float64",
		"Int64":  "int64",
		"Str":    "string",
		"Link":   "trace.Trace",
		"Any":    "interface{}",
		"Time":   "time.Time",
		"Enum":   "xopconst.Enum",
	},
	"IntsPlus": {
		"Int":      "int",
		"Int8":     "int8",
		"Int16":    "int16",
		"Int32":    "int32",
		"Int64":    "int64",
		"Duration": "time.Duration",
	},
	"Ints": {
		"Int":   "int",
		"Int8":  "int8",
		"Int16": "int16",
		"Int32": "int32",
		"Int64": "int64",
	},
	"Uints": {
		"Uint":   "uint",
		"Uint8":  "uint8",
		"Uint16": "uint16",
		"Uint32": "uint32",
	},
	"BaseData": {
		"Int":      "int64",
		"Uint":     "uint64",
		"Str":      "string",
		"Bool":     "bool",
		"Any":      "interface{}",
		"Link":     "trace.Trace",
		"Error":    "error",
		"Time":     "time.Time",
		"Duration": "time.Duration",
	},
	"AllData": {
		"Int":      "int",
		"Int8":     "int8",
		"Int16":    "int16",
		"Int32":    "int32",
		"Int64":    "int64",
		"Uint":     "uint",
		"Uint8":    "uint8",
		"Uint16":   "uint16",
		"Uint32":   "uint32",
		"Uint64":   "uint64",
		"Str":      "string",
		"Bool":     "bool",
		"Any":      "interface{}",
		"Link":     "trace.Trace",
		"Duration": "time.Duration",
		"Error":    "error",
	},
	"HexBytes": {
		"HexBytes1":  "1",
		"HexBytes8":  "8",
		"HexBytes16": "16",
	},
}

var allLines []string
var index int
var currentPackage string

func main() {
	fmt.Println("// This file is generated, DO NOT EDIT.  It comes from the corresponding .zzzgo file")
	fmt.Println("") // prevent above comment from becoming a package comment
	var reader = bufio.NewReader(os.Stdin)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		allLines = append(allLines, line)
	}

	for index = 0; index < len(allLines); index++ {
		line := allLines[index]
		if m := macroRE.FindStringSubmatch(line); m != nil {
			macroExpand(m[1], m[2], m[3])
			continue
		}
		if m := packageRE.FindStringSubmatch(line); m != nil {
			currentPackage = m[1]
		}
		fmt.Print(line)
	}
}

func macroExpand(indent string, macro string, skipList string) {
	m, ok := macros[macro]
	if !ok {
		panic(fmt.Errorf("'%s' isn't a valid macro, at line %d", macro, index+1))
	}
	var lines []string
	for index++; index < len(allLines); index++ {
		line := allLines[index]
		i := indentRE.FindStringSubmatch(line)
		if i == nil {
			panic(fmt.Errorf("indent RE did not match on line %d: '%s'", index+1, line))
		}
		if (indent != "" && len(i[1]) < len(indent)) || line == "\n" || line == "\r\n" {
			index--
			break
		}
		if macroRE.MatchString(line) {
			index--
			break
		}
		lines = append(lines, line)
	}

	skips := make(map[string]struct{})
	for _, skip := range strings.Split(skipList, ",") {
		skips[skip] = struct{}{}
	}

	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, name := range keys {
		if _, ok := skips[name]; ok {
			continue
		}
		typ := m[name]
		if currentPackage != "" {
			typ = strings.TrimPrefix(typ, currentPackage+".")
		}
		replMap := map[string]string{
			"ZZZ": name,
			"zzz": typ,
		}
		for _, line := range lines {
			rewritten := zzzRE.ReplaceAllStringFunc(line, func(s string) string {
				return replMap[s]
			})
			fmt.Print(rewritten)
		}
	}
}
