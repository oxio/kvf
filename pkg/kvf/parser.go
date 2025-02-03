package kvf

import (
	"fmt"
	"strings"
)

type LineParser struct {
}

func NewLineParser() *LineParser {
	return &LineParser{}
}

func (*LineParser) Parse(line string) (*Item, error) {
	line = strings.TrimSpace(line)

	item := &Item{}

	if line == "" {
		return &Item{IsEmpty: true}, nil
	} else if strings.HasPrefix(line, "#") {
		item.IsComment = true
		item.Val = strings.TrimSpace(strings.TrimLeft(line, "#"))
	} else {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid line: %s", line)
		}
		item.Key = strings.TrimSpace(parts[0])
		item.Val = strings.TrimSpace(parts[1])

		if strings.HasPrefix(item.Val, "\"") && strings.HasSuffix(item.Val, "\"") {
			item.Quote = "\""
		} else if strings.HasPrefix(item.Val, "'") && strings.HasSuffix(item.Val, "'") {
			item.Quote = "'"
		}
	}

	if item.Quote != "" {
		item.Val = item.Val[1 : len(item.Val)-1]
	}

	return item, nil
}
