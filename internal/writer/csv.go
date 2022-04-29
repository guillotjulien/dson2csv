package writer

import (
	"fmt"
	"reflect"
	"unicode/utf8"

	"golang.org/x/exp/slices"
)

const MAX_CHAR_PER_CELL = 32767
const LEEWAY = 1000 // give us enough to ignore how excel count (code point or runes)

type CSVOutput struct {
	Headers []string
	Data    [][]string
}

func MapToCSV(rows []map[string]string) CSVOutput {
	headersMap := make(map[string]bool)

	for _, row := range rows {
		for k := range row {
			if !headersMap[k] {
				headersMap[k] = true
			}
		}
	}

	o := CSVOutput{}

	for _, h := range reflect.ValueOf(headersMap).MapKeys() {
		o.Headers = append(o.Headers, h.Interface().(string))
	}
	slices.Sort(o.Headers)

	for _, row := range rows {
		var r []string
		for _, h := range o.Headers {
			if utf8.RuneCountInString(row[h]) > MAX_CHAR_PER_CELL { // Excel only allow a certain number of characters per cells
				row[h] = fmt.Sprintf("%s...", string([]rune(row[h])[:MAX_CHAR_PER_CELL-LEEWAY]))
			}

			r = append(r, row[h])
		}
		o.Data = append(o.Data, r)
	}

	return o
}
