package main

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/scanner"

	"github.com/guillotjulien/dson2csv/internal"
	"github.com/guillotjulien/dson2csv/internal/writer"
)

// These values are stored in the parseState stack.
// They give the current state of a composite value
// being scanned. If the parser is inside a nested value
// the parseState describes the nested state, outermost at entry 0.
const (
	parseObjectKey   = "key"    // parsing object key (before colon)
	parseObjectValue = "object" // parsing object value (after colon)
	parseArrayValue  = "array"  // parsing array value
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		log.Fatal("Missing argument for path. Usage: bson-to-csv <input_path>")
	}

	f, err := os.Open(args[0])
	if err != nil {
		log.Fatal("Failed to read input file: ", err)
	}
	defer f.Close()

	// f, err := ioutil.ReadFile(args[0])
	// if err != nil {
	// 	log.Fatal("Failed to read input file: ", err)
	// }

	var s scanner.Scanner
	s.Init(bufio.NewReader(f)) // FIXME: why does using bufio is a lot slower than reading as a byte array?
	// s.Init(bytes.NewReader(f))

	rows := make([]map[string]string, 0)

	for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() { // TODO: extract the content of the loop to a scanner so that we can invoke it using unit tests / fuzzing
		token := s.TokenText()

		if token == "{" {
			row, err := consumeObject(s)
			if err != nil {
				log.Fatalf("[%v:%v] Unexpected error %v", s.Position.Line, s.Position.Column, err)
			}

			rows = append(rows, row)
		}
	}

	// fmt.Println(rows)

	o := writer.MapToCSV(rows)

	fileName := strings.TrimSuffix(args[0], filepath.Ext(args[0]))
	of, err := os.Create(fmt.Sprintf("%s.csv", fileName))
	if err != nil {
		log.Fatalf("Failed to create output file: Error: %v\n", err)
	}
	defer of.Close()

	// When encoded inside of WSL, excel break character encoding when opened in Windows
	// https://forum.golangbridge.org/t/csv-characters-problem-with-write/11625/3
	w := csv.NewWriter(of)
	headers := o.Headers
	values := o.Data
	if err := w.Write(headers); err != nil {
		log.Fatalf("Failed to write CSV headers. Error: %v\n", err)
	}

	for _, r := range values {
		if err := w.Write(r); err != nil {
			log.Fatalf("Failed to write CSV row: %v. Error: %v\n", r, err)
		}
	}

	w.Flush()
}

func consumeObject(s scanner.Scanner) (val map[string]string, err error) {
	if s.TokenText() != "{" {
		return nil, errors.New("called consumeObject on non-object structure")
	}

	var prevToken string

	fieldPath := make(internal.Stack, 0)

	state := make(internal.Stack, 0)
	state = state.Push(parseObjectValue)

	val = make(map[string]string)

	for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
		if len(state) == 0 {
			break // we successfully consumed the object
		}

		token := s.TokenText()

		if isIgnoreToken(token) { // we don't want to collect those
			continue
		}

		// Don't rely on "," so that we can handle standard and relaxed JSON as well
		switch token {
		case "{":
			state = state.Push(parseObjectValue)
		case "}":
			if state, err = state.Pop(); err != nil { // pop the object
				return nil, errors.New("trying to close object before one was opened")
			}
			if prev := state.Peek(); prev == parseObjectKey { // pop the key
				if state, err = state.Pop(); err != nil {
					return nil, errors.New("structure does not respect key:value format")
				}
				if fieldPath, err = fieldPath.Pop(); err != nil {
					return nil, errors.New("structure does not respect key:value format")
				}
			}
		case "[":
			// TODO: if prev state isn't parseObjectKey, then the structure is incorrect (e.g. "{[]}")
			token, err = consumeArray(s)
		case "]":
			return nil, errors.New("trying to close array before one was opened")
		}

		if token == ":" {
			fieldPath = fieldPath.Push(strings.Trim(strings.Trim(prevToken, `"`), `'`))
			state = state.Push(parseObjectKey)
		}

		prev := state.Peek()
		if prev == parseObjectKey && prevToken == ":" {
			val[strings.Join(fieldPath, ".")] = strings.Trim(strings.Trim(token, `"`), `'`)

			if fieldPath, err = fieldPath.Pop(); err != nil {
				return nil, errors.New("structure does not respect key:value format")
			}
			if state, err = state.Pop(); err != nil {
				return nil, errors.New("structure does not respect key:value format")
			}
		}

		prevToken = token
	}

	if len(state) > 0 {
		return nil, errors.New("reached end of file before closing object")
	}

	return
}

func consumeArray(s scanner.Scanner) (val string, err error) {
	if s.TokenText() != "[" {
		return "", errors.New("called consumeArray on non-array structure")
	}

	tokens := make(internal.Stack, 0)
	tokens = tokens.Push(s.TokenText())

	state := make(internal.Stack, 0)
	state = state.Push(parseArrayValue)

	for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
		if len(state) == 0 {
			break // we successfully consumed the array
		}

		token := s.TokenText()

		if isIgnoreToken(token) { // we don't want to collect those
			continue
		}

		switch token {
		case "{":
			state = state.Push(parseObjectValue)
		case "}":
			if state, err = state.Pop(); err != nil {
				return "", errors.New("trying to close object before one was opened")
			}
		case "[":
			state = state.Push(parseArrayValue)
		case "]":
			if state, err = state.Pop(); err != nil {
				return "", errors.New("trying to close array before one was opened")
			}
		}

		tokens = tokens.Push(token)
	}

	if len(state) > 0 {
		return "", errors.New("reached end of file before closing array")
	}

	return strings.Join(tokens, " "), nil
}

func isIgnoreToken(tok string) bool {
	switch tok {
	case "ObjectId",
		"ISODate",
		"Date",
		"Timestamp",
		"Int32",
		"Long",
		"Decimal128",
		"NumberLong",
		"NumberInt",
		"NumberDecimal",
		"(",
		")":
		return true
	}

	return false
}
