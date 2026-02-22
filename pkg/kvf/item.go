package kvf

import (
	"fmt"
	"strings"
)

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
	if "" == key {
		return nil, fmt.Errorf("key is empty")
	}
	item := &Item{
		Key: key,
		Val: val,
	}
	// If value contains newlines, wrap it in double quotes
	if strings.Contains(val, "\n") {
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
	return fmt.Sprintf("%s=%s%s%s\n", i.Key, i.Quote, i.Val, i.Quote)
}
