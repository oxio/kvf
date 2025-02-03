package kvf

import "fmt"

type Item struct {
	IsEmpty   bool
	IsComment bool
	Key       string
	Val       string
	Quote     string
}

func NewItem(key string, val string) (*Item, error) {
	if "" == key {
		return nil, fmt.Errorf("key is empty")
	}
	return &Item{
		Key: key,
		Val: val,
	}, nil
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
