package writer

import "reflect"

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

	// FIXME: when present, _id or id should be the first field
	for _, h := range reflect.ValueOf(headersMap).MapKeys() {
		o.Headers = append(o.Headers, h.Interface().(string))
	}

	for _, row := range rows {
		var r []string
		for _, h := range o.Headers {
			if row[h] != "" {
				r = append(r, row[h])
			}
		}
		o.Data = append(o.Data, r)
	}

	return o
}
