package kvf

import (
	"fmt"
	"strings"
)

// escapeValue escapes backslashes and the specified quote character in a value
func escapeValue(val string, quote string) string {
	// First escape backslashes, then escape the quote char
	result := strings.ReplaceAll(val, "\\", "\\\\")
	if quote == "\"" {
		result = strings.ReplaceAll(result, "\"", "\\\"")
	} else {
		result = strings.ReplaceAll(result, "'", "\\'")
	}
	return result
}

// ParseState represents the parsing state of an Item
type ParseState int

const (
	// ParseStateComplete indicates the item is fully parsed
	ParseStateComplete ParseState = iota
	// ParseStateMultiLineStart indicates a multi-line value has started but not completed
	ParseStateMultiLineStart
	// ParseStateMultiLineContinue indicates a multi-line value is continuing
	ParseStateMultiLineContinue
)

type Item struct {
	IsEmpty    bool
	IsComment  bool
	Key        string
	Val        string
	Quote      string
	ParseState ParseState
}

func NewItem(key string, val string) (*Item, error) {
	if key == "" {
		return nil, fmt.Errorf("key is empty")
	}
	item := &Item{
		Key: key,
		Val: val,
	}

	hasDoubleQuote := strings.Contains(val, "\"")
	hasSingleQuote := strings.Contains(val, "'")
	hasNewline := strings.Contains(val, "\n")
	hasBackslash := strings.Contains(val, "\\")

	// Determine if quoting is needed and which quote to use
	if hasNewline || hasBackslash || (hasDoubleQuote && hasSingleQuote) {
		// Must use double quotes, will need escaping
		item.Quote = "\""
	} else if hasDoubleQuote {
		// Use single quotes to avoid escaping
		item.Quote = "'"
	} else if hasSingleQuote {
		// Use double quotes to avoid escaping
		item.Quote = "\""
	}

	return item, nil
}

type ItemCollection struct {
	Items *[]*Item
}

func NewItemCollection() *ItemCollection {
	return &ItemCollection{
		Items: &[]*Item{},
	}
}

func (ic *ItemCollection) Add(item *Item) {
	*ic.Items = append(*ic.Items, item)
}

func (i *Item) ToLine() string {
	if i.IsEmpty {
		return "\n"
	}
	if i.IsComment {
		return fmt.Sprintf("# %s\n", i.Val)
	}

	if i.Quote == "" {
		return fmt.Sprintf("%s=%s\n", i.Key, i.Val)
	}

	// Escape the value for quoted output
	escapedVal := escapeValue(i.Val, i.Quote)
	return fmt.Sprintf("%s=%s%s%s\n", i.Key, i.Quote, escapedVal, i.Quote)
}
