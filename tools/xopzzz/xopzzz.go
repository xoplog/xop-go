package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var macroRE = regexp.MustCompile(`^(\s*)//\s?MACRO (\w+)(?:\s+(SKIP|ONLY):(\S+))?\s*$`)
var errorRE = regexp.MustCompile(`^(\s*)//MACRO/`)
var indentRE = regexp.MustCompile(`^(\s*)(?:\S|$)`)
var zzzRE = regexp.MustCompile(`(zzz|Zzz|ZZZ|zZZ)`)
var packageRE = regexp.MustCompile(`^package (\w+)`)
var conditionalRE = regexp.MustCompile(`^\s*//\s?CONDITIONAL (?:(?:ONLY:(\S+))|(?:SKIP:(\S+)))\s*$`)
var endConditionalRE = regexp.MustCompile(`^\s*//\s?END CONDITIONAL\s*$`)
var elseConditionalRE = regexp.MustCompile(`^\s*//\s?ELSE CONDITIONAL\s*$`)

var macros = map[string]map[string]string{
	// ZZZAttribute are the span-attributes in the main logger.  These
	// turn into Span methods like Span.BoolAttribute() and a corresponding
	// xopat.BoolAttribute type.
	//
	// Also change ../xopproto/ingest.proto when changing these
	//
	"ZZZAttribute": {
		"Bool":     "bool",
		"Float64":  "float64",
		"Int64":    "int64",
		"Int32":    "int32",
		"Int16":    "int16",
		"Int8":     "int8",
		"Int":      "int",
		"Link":     "xoptrace.Trace",
		"String":   "string",
		"Any":      "xopbase.ModelArg",
		"Time":     "time.Time",
		"Duration": "time.Duration",
		"Enum":     "xopat.Enum",
	},
	// BaseAttributes are the span attributes that base loggers must
	// implement.  These turn into things like Base.MetadataBool()
	"BaseAttribute": {
		"Bool":    "bool",
		"Float64": "float64",
		"Int64":   "int64",
		"String":  "string",
		"Link":    "xoptrace.Trace",
		"Any":     "xopbase.ModelArg",
		"Time":    "time.Time",
		"Enum":    "xopat.Enum",
	},
	"BaseAttributeExample": {
		"Bool":    "true",
		"Float64": "float64(0.0)",
		"String":  "\"\"",
		"Link":    "xoptrace.Trace{}",
		"Any":     "interface{}",
		"Time":    "time.Time{}",
		"Enum":    "xopat.Enum",
		"Int64":   "int64(0)",
	},
	"IntsPlus": {
		"Int":      "int",
		"Int8":     "int8",
		"Int16":    "int16",
		"Int32":    "int32",
		"Int64":    "int64",
		"Duration": "time.Duration",
	},
	"SimpleAttributeReconstructionPB": {
		"Bool":    "",
		"Float64": "v.FloatValue",
		"String":  "v.StringValue",
		"Link":    "",
		"Any":     "",
		"Time":    "time.Unix(0,v.IntValue)",
		"Enum":    "",
		"Int64":   "v.IntValue",
	},
	"SimpleAttributeReconstructionJSON": {
		"Bool":    "bool",
		"Float64": "v.FloatValue",
		"String":  "v.StringValue",
		"Link":    "",
		"Any":     "",
		"Time":    "time.Unix(0,v.IntValue)",
		"Enum":    "",
		"Int64":   "v.IntValue",
	},
	"Ints": {
		"Int":   "int",
		"Int8":  "int8",
		"Int16": "int16",
		"Int32": "int32",
		"Int64": "int64",
	},
	"Uints": {
		"Uint":    "uint",
		"Uint8":   "uint8",
		"Uint16":  "uint16",
		"Uint32":  "uint32",
		"Uint64":  "uint64",
		"Uintptr": "uintptr",
	},
	// BaseData are the data types that are supported on a per-line basis in xopbase.Line
	// Note: Enum is not included since it needs special handling every time
	"BaseData": {
		"Int64":    "int64",
		"Uint64":   "uint64",
		"String":   "string",
		"Bool":     "bool",
		"Any":      "xopbase.ModelArg",
		"Time":     "time.Time",
		"Duration": "time.Duration",
		"Float64":  "float64",
	},
	"BaseDataWithoutType": {
		"Bool":     "bool",
		"Any":      "xopbase.ModelArg",
		"Time":     "time.Time",
		"Duration": "time.Duration",
	},
	"BaseDataWithType": {
		"Int64":   "int64",
		"Uint64":  "uint64",
		"Float64": "float64",
		"String":  "string",
	},
	"LineEndersWithData": {
		"Link":  "xoptrace.Trace",
		"Model": "xopbase.ModelArg",
	},
	"LineEndersWithoutData": {
		"Msg":      "string",
		"Template": "string",
	},
	// AllData includes all all span metadata, all base types, all line types
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
		"Uintptr":  "uintptr",
		"String":   "string",
		"Bool":     "bool",
		"Any":      "interface{}",
		"Duration": "time.Duration",
		"Error":    "error",
		"Float64":  "float64",
		"Float32":  "float32",
		"Time":     "time.Time",
		"Stringer": "fmt.Stringer",
		"Link":     "xoptrace.Trace",
	},
	"HexBytes": {
		"HexBytes1":  "1",
		"HexBytes8":  "8",
		"HexBytes16": "16",
	},
	// Enumer is all of the generated enumers, used for generating a test
	"Enumer": {
		"EventType": "xoprecorder",
		"Level":     "xopnum",
	},
	"OTELAttributes": {
		"String":       "string",
		"Int64":        "int64",
		"Float64":      "float64",
		"Bool":         "bool",
		"StringSlice":  "[]string",
		"Int64Slice":   "[]int64",
		"Float64Slice": "[]float64",
		"BoolSlice":    "[]bool",
		"Stringer":     "fmt.Stringer",
	},
	"OTELTypes": {
		"STRING":  "string",
		"INT64":   "int64",
		"FLOAT64": "float64",
		"BOOL":    "bool",
	},
	"OTELSpanKinds": {
		// "SpanKindUnspecified": "false", omitted because it's not valid
		"SpanKindInternal": "false",
		"SpanKindServer":   "true",
		"SpanKindClient":   "true",
		"SpanKindProducer": "true",
		"SpanKindConsumer": "true",
	},
	// Note: these map to base types, not exact types.  Exact types
	// are next.
	"DataTypeAbbreviations": {
		"i":        "Int64",
		"i8":       "Int64",
		"i16":      "Int64",
		"i32":      "Int64",
		"i64":      "Int64",
		"u":        "Uint64",
		"u8":       "Uint64",
		"u16":      "Uint64",
		"u32":      "Uint64",
		"u64":      "Uint64",
		"uintptr":  "Uint64",
		"f32":      "Float64",
		"f64":      "Float64",
		"any":      "Any",
		"bool":     "Bool",
		"dur":      "Duration",
		"time":     "Time",
		"s":        "String",
		"stringer": "String",
		"enum":     "Enum",
		"error":    "String",
	},
	"DataTypeAbbreviationsExact": {
		"i":        "Int",
		"i8":       "Int8",
		"i16":      "Int16",
		"i32":      "Int32",
		"i64":      "Int64",
		"u":        "Uint",
		"u8":       "Uint8",
		"u16":      "Uint16",
		"u32":      "Uint32",
		"u64":      "Uint64",
		"uintptr":  "Uintptr",
		"f32":      "Float32",
		"f64":      "Float64",
		"any":      "Any",
		"bool":     "Bool",
		"dur":      "Duration",
		"time":     "Time",
		"s":        "String",
		"stringer": "Stringer",
		"enum":     "Enum",
		"error":    "Error",
	},
	"LogLevel": {
		"Trace": "",
		"Debug": "",
		"Info":  "",
		"Log":   "",
		"Warn":  "",
		"Error": "",
		"Alert": "",
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
		if len(allLines) == 0 && line == "// TEMPLATE-FILE\n" {
			continue
		}
		allLines = append(allLines, line)
	}

	for index = 0; index < len(allLines); index++ {
		line := allLines[index]
		if m := macroRE.FindStringSubmatch(line); m != nil {
			macroExpand(m[1], m[2], m[3] == "SKIP", m[4])
			continue
		}
		if m := packageRE.FindStringSubmatch(line); m != nil {
			currentPackage = m[1]

			for _, macros := range macros {
				for k, v := range macros {
					macros[k] = strings.TrimPrefix(v, currentPackage+".")
				}
			}
		}
		if errorRE.MatchString(line) {
			panic(fmt.Errorf("found invalid //MACRO at line %d", index+1))
		}
		fmt.Print(line)
	}
}

var toTitle = cases.Title(language.Und)

func macroExpand(indent string, macro string, skip bool, skipList string) {
	values, ok := macros[macro]
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
	if skipList != "" {
		for _, skip := range strings.Split(skipList, ",") {
			skips[skip] = struct{}{}
		}
	}

	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, name := range keys {
		_, ok := skips[name]
		if len(skips) > 0 && skip == ok {
			continue
		}
		typ := values[name]
		if currentPackage != "" {
			typ = strings.TrimPrefix(typ, currentPackage+".")
		}
		replMap := map[string]string{
			"ZZZ": name,
			"zzz": typ,
			"Zzz": toTitle.String(typ),
			"zZZ": strings.ToLower(name),
		}
		var skipping bool
		for _, line := range lines {
			if m := conditionalRE.FindStringSubmatch(line); m != nil {
				if m[1] != "" {
					// ONLY
					skipping = true
					for _, n := range strings.Split(m[1], ",") {
						if _, ok := values[n]; !ok {
							panic(fmt.Errorf("value %s is not part of %s", n, macro))
						}
						if n == name {
							skipping = false
							break
						}
					}
				} else {
					// SKIP
					skipping = false
					for _, n := range strings.Split(m[2], ",") {
						if _, ok := values[n]; !ok {
							panic(fmt.Errorf("value %s is not part of %s", n, macro))
						}
						if n == name {
							skipping = true
							break
						}
					}
				}
				continue
			} else if elseConditionalRE.MatchString(line) {
				skipping = !skipping
				continue
			} else if endConditionalRE.MatchString(line) {
				skipping = false
				continue
			}
			if skipping {
				continue
			}
			rewritten := zzzRE.ReplaceAllStringFunc(line, func(s string) string {
				return replMap[s]
			})
			fmt.Print(rewritten)
		}
		skipping = false
	}
}
