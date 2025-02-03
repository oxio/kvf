package kvf

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/oxio/kvf/internal/fileop"
)

type Repo interface {
	Get(key string) (*Item, error)
	Set(item *Item) error
}

var _ Repo = &FileRepo{}

var (
	ErrItemNotFound = errors.New("item not found")
)

type FileRepo struct {
	adapter fileop.FileAdapter
	parser  *LineParser
}

func NewFileRepo(filePath string, noErrorOnInaccessibleFile bool) *FileRepo {
	return &FileRepo{
		adapter: fileop.NewFileAdapter(filePath, noErrorOnInaccessibleFile),
		parser:  NewLineParser(),
	}
}

func (r *FileRepo) FindAll() (*[]*Item, error) {
	var collection = NewItemCollection()
	err := r.adapter.ReadByLine(r.makeReader(collection))
	if err != nil {
		return nil, err
	}
	return collection.Items, nil
}

func (r *FileRepo) Get(key string) (*Item, error) {
	if "" == key {
		return nil, fmt.Errorf("key is empty")
	}

	var collection = NewItemCollection()
	err := r.adapter.ReadByLine(r.makeReader(collection))
	if err != nil {
		return nil, err
	}

	for _, item := range *collection.Items {
		if item.IsEmpty || item.IsComment {
			continue
		}
		if item.Key == key {
			return item, nil
		}
	}

	return nil, ErrItemNotFound
}

func (r *FileRepo) Set(item *Item) error {
	var collection = NewItemCollection()
	read := r.makeReader(collection)
	update := r.makeUpdater(collection, item)
	write := r.makeWriter(collection)

	return r.adapter.EnsureUpdate(read, update, write)
}

func (r *FileRepo) makeReader(collection *ItemCollection) fileop.ReaderFunc {
	return func(line string) error {
		item, err := r.parser.Parse(line)
		if err != nil {
			return err
		}
		*collection.Items = append(*collection.Items, item)
		return nil
	}
}

func (r *FileRepo) makeUpdater(collection *ItemCollection, incoming *Item) fileop.UpdateFunc {
	return func() error {
		found := false
		for k, item := range *collection.Items {
			if item.Key == incoming.Key {
				item.Val = incoming.Val
				(*collection.Items)[k] = item
				found = true
				break
			}
		}
		if !found {
			*collection.Items = append(*collection.Items, incoming)
		}
		return nil
	}
}

func (r *FileRepo) makeWriter(collection *ItemCollection) fileop.WriterFunc {
	return func(writer *bufio.Writer) (bytesWritten int64, err error) {
		for _, item := range *collection.Items {
			var n, err = writer.WriteString(item.ToLine())
			bytesWritten += int64(n)
			if err != nil {
				return bytesWritten, err
			}
		}
		return bytesWritten, nil
	}
}
