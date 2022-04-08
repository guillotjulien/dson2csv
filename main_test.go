package main

import (
	"fmt"
	"strings"
	"testing"
	"text/scanner"
)

func TestConsumeObject(t *testing.T) {
	data := `{
		_id : ObjectId("5099803df3f4948bd2f98391"),
		name : {
			first: "Alan",
			last: "Turing"
		},
		birth : ISODate('Jun 23, 1912'),
		death : ISODate('Jun 07, 1954'),
		views : NumberLong(1250000),
	}`

	want := map[string]string{
		"_id":        "5099803df3f4948bd2f98391",
		"name.first": "Alan",
		"name.last":  "Turing",
		"birth":      "Jun 23, 1912",
		"death":      "Jun 07, 1954",
		"views":      "1250000",
	}

	var s scanner.Scanner
	s.Init(strings.NewReader(data))
	s.Scan()

	got, err := consumeObject(&s)
	if err != nil {
		t.Fatalf("should return a map. got %v", err)
	}

	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestFlatObject(t *testing.T) {
	data := `{ name : { first: "Alan", last: "Turing" } }`

	want := map[string]string{
		"name.first": "Alan",
		"name.last":  "Turing",
	}

	var s scanner.Scanner
	s.Init(strings.NewReader(data))
	s.Scan()

	got, err := consumeObject(&s)
	if err != nil {
		t.Fatalf("should return a map. got %v", err)
	}

	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestConsumeArraySimple(t *testing.T) {
	data := `[
		1,
		2,
		3
	]`

	want := "[ 1 , 2 , 3 ]"

	var s scanner.Scanner
	s.Init(strings.NewReader(data))
	s.Scan()

	got, err := consumeArray(&s)
	if err != nil {
		t.Fatalf("should return a string. got %v", err)
	}

	if got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestConsumeArrayObject(t *testing.T) {
	data := `[
		{
			name: {
				first: "Harry",
				last: "Potter"
			}
		},
		{
			name: {
				first: "Hermione",
				last: "Granger"
			}
		}
	]`

	want := "[ { name : { first : \"Harry\" , last : \"Potter\" } } , { name : { first : \"Hermione\" , last : \"Granger\" } } ]"

	var s scanner.Scanner
	s.Init(strings.NewReader(data))
	s.Scan()

	got, err := consumeArray(&s)
	if err != nil {
		t.Fatalf("should return a string. got %v", err)
	}

	if got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestConsumeArrayFlat(t *testing.T) {
	data := `[1, 2, 3]`

	want := "[ 1 , 2 , 3 ]"

	var s scanner.Scanner
	s.Init(strings.NewReader(data))
	s.Scan()

	got, err := consumeArray(&s)
	if err != nil {
		t.Fatalf("should return a string. got %v", err)
	}

	if got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}
