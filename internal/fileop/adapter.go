package fileop

import (
	"bufio"
	"github.com/oxio/kvf/v2/internal/lock"
	"io"
	"os"
)

type FileAdapter interface {
	ReadByLine(lineCallback ReaderFunc) error
	EnsureReadByLine(lineCallback ReaderFunc) error
	EnsureUpdate(
		readCallback ReaderFunc,
		updateCallback UpdateFunc,
		writeCallback WriterFunc,
	) error
}

type ReaderFunc func(line string) error
type UpdateFunc func() error
type WriterFunc func(writer *bufio.Writer) (bytesWritten int64, err error)

type DefaultAdapter struct {
	filePath                string
	noErrOnInaccessibleFile bool
}

var _ FileAdapter = &DefaultAdapter{}

func NewFileAdapter(filePath string, noErrOnInaccessibleFile bool) *DefaultAdapter {
	return &DefaultAdapter{
		filePath:                filePath,
		noErrOnInaccessibleFile: noErrOnInaccessibleFile,
	}
}

func (adapter *DefaultAdapter) ReadByLine(lineCallback ReaderFunc) error {
	fp, err := os.Open(adapter.filePath)
	if err != nil {
		if adapter.noErrOnInaccessibleFile && (os.IsNotExist(err) || os.IsPermission(err)) {
			return nil
		}
		return err
	}
	defer func(file *os.File) {
		err = file.Close()
	}(fp)

	l, err := lock.New(fp.Name())
	if err != nil {
		return err
	}
	defer func() {
		err = l.Release()
	}()

	err = adapter.readLine(fp, lineCallback)

	return err
}

func (adapter *DefaultAdapter) EnsureReadByLine(lineCallback ReaderFunc) error {
	fp, err := os.OpenFile(adapter.filePath, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err = file.Close()
	}(fp)

	l, err := lock.New(fp.Name())
	if err != nil {
		return err
	}
	defer func() {
		err = l.Release()
	}()

	err = adapter.readLine(fp, lineCallback)

	return err
}

func (adapter *DefaultAdapter) readLine(fp *os.File, lineCallback ReaderFunc) (err error) {
	scanner := bufio.NewScanner(fp)
	for scanner.Scan() {
		line := scanner.Text()
		err = lineCallback(line)
		if err != nil {
			return err
		}
	}

	return err
}

func (adapter *DefaultAdapter) EnsureUpdate(
	readCallback ReaderFunc,
	updateCallback UpdateFunc,
	writeCallback WriterFunc,
) error {
	l, err := lock.New(adapter.filePath)
	if err != nil {
		return err
	}
	defer func() {
		err = l.Release()
	}()

	fp, err := os.OpenFile(adapter.filePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err = file.Close()
	}(fp)

	writer := bufio.NewWriter(fp)
	defer func(writer *bufio.Writer) {
		err = writer.Flush()
	}(writer)

	err = adapter.readLine(fp, readCallback)
	if err != nil {
		return err
	}

	err = updateCallback()
	if err != nil {
		return err
	}

	_, err = fp.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	bytesWritten, err := writeCallback(writer)
	if err != nil {
		return err
	}
	err = fp.Truncate(bytesWritten)

	return err
}
