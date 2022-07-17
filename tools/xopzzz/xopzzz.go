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

var macros = map[string]map[string]string{
	"ZZZAttribute": {
		"Bool":     "bool",
		"Int":      "int",
		"Str":      "string",
		"Link":     "trace.Trace",
		"Any":      "interface{}",
		"Time":     "time.Time",
		"Duration": "time.Duration",
	},
}

var allLines []string
var index int

func main() {
	fmt.Println("// This file is generated, DO NOT EDIT")
	fmt.Println("// It is generated from the corresponding .zzzgo file using zopzzz")
	fmt.Println("//")
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
		replMap := map[string]string{
			"ZZZ": name,
			"zzz": m[name],
		}
		for _, line := range lines {
			rewritten := zzzRE.ReplaceAllStringFunc(line, func(s string) string {
				return replMap[s]
			})
			fmt.Print(rewritten)
		}
	}
}
