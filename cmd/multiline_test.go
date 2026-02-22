package cmd

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetMultiLineValue_DoubleQuote(t *testing.T) {
	content := `KEY1="line1
line2
line3"
KEY2=simple
`
	tmpFile, err := createRandomTestFileWithContent(content)
	if err != nil {
		t.Fatalf("Error setting up test file: %v", err)
	}
	defer removeTestFile(tmpFile.Name())

	cmd := newGetCmd()
	outBuff := bytes.NewBufferString("")
	cmd.SetOut(outBuff)
	cmd.SetArgs([]string{tmpFile.Name(), "KEY1"})

	err = cmd.Execute()
	assert.NoError(t, err)

	outContent, err := io.ReadAll(outBuff)
	assert.NoError(t, err)
	assert.Equal(t, "line1\nline2\nline3", string(outContent))
}

func TestGetMultiLineValue_SingleQuote(t *testing.T) {
	content := `KEY1='line1
line2
line3'
KEY2=simple
`
	tmpFile, err := createRandomTestFileWithContent(content)
	if err != nil {
		t.Fatalf("Error setting up test file: %v", err)
	}
	defer removeTestFile(tmpFile.Name())

	cmd := newGetCmd()
	outBuff := bytes.NewBufferString("")
	cmd.SetOut(outBuff)
	cmd.SetArgs([]string{tmpFile.Name(), "KEY1"})

	err = cmd.Execute()
	assert.NoError(t, err)

	outContent, err := io.ReadAll(outBuff)
	assert.NoError(t, err)
	assert.Equal(t, "line1\nline2\nline3", string(outContent))
}

func TestGetMultiLineValue_WithEmptyLine(t *testing.T) {
	content := `KEY1="line1

line3"
`
	tmpFile, err := createRandomTestFileWithContent(content)
	if err != nil {
		t.Fatalf("Error setting up test file: %v", err)
	}
	defer removeTestFile(tmpFile.Name())

	cmd := newGetCmd()
	outBuff := bytes.NewBufferString("")
	cmd.SetOut(outBuff)
	cmd.SetArgs([]string{tmpFile.Name(), "KEY1"})

	err = cmd.Execute()
	assert.NoError(t, err)

	outContent, err := io.ReadAll(outBuff)
	assert.NoError(t, err)
	assert.Equal(t, "line1\n\nline3", string(outContent))
}

func TestSetMultiLineValue(t *testing.T) {
	filePath := getRandomTestFilePath()
	defer removeTestFile(filePath)

	cmd := newSetCmd()
	cmd.SetArgs([]string{filePath, "KEY", "line1\nline2\nline3"})

	err := cmd.Execute()
	assert.NoError(t, err)

	// Verify file content
	assertFileContentEquals(t, filePath, "KEY=\"line1\nline2\nline3\"\n")

	// Now get the value back
	getCmd := newGetCmd()
	outBuff := bytes.NewBufferString("")
	getCmd.SetOut(outBuff)
	getCmd.SetArgs([]string{filePath, "KEY"})

	err = getCmd.Execute()
	assert.NoError(t, err)

	outContent, err := io.ReadAll(outBuff)
	assert.NoError(t, err)
	assert.Equal(t, "line1\nline2\nline3", string(outContent))
}

func TestMultiLineRoundTrip(t *testing.T) {
	filePath := getRandomTestFilePath()
	defer removeTestFile(filePath)

	multiLineValue := "First line\nSecond line\nThird line"

	// Set the multi-line value
	setCmd := newSetCmd()
	setCmd.SetArgs([]string{filePath, "MULTI_KEY", multiLineValue})
	err := setCmd.Execute()
	assert.NoError(t, err)

	// Get the value back
	getCmd := newGetCmd()
	outBuff := bytes.NewBufferString("")
	getCmd.SetOut(outBuff)
	getCmd.SetArgs([]string{filePath, "MULTI_KEY"})
	err = getCmd.Execute()
	assert.NoError(t, err)

	outContent, err := io.ReadAll(outBuff)
	assert.NoError(t, err)
	assert.Equal(t, multiLineValue, string(outContent))
}

func TestMultiLineWithOtherItems(t *testing.T) {
	content := `KEY1=simple
KEY2="multi
line"
KEY3=another
`
	tmpFile, err := createRandomTestFileWithContent(content)
	if err != nil {
		t.Fatalf("Error setting up test file: %v", err)
	}
	defer removeTestFile(tmpFile.Name())

	// Get KEY1
	cmd1 := newGetCmd()
	outBuff1 := bytes.NewBufferString("")
	cmd1.SetOut(outBuff1)
	cmd1.SetArgs([]string{tmpFile.Name(), "KEY1"})
	err = cmd1.Execute()
	assert.NoError(t, err)
	out1, _ := io.ReadAll(outBuff1)
	assert.Equal(t, "simple", string(out1))

	// Get KEY2
	cmd2 := newGetCmd()
	outBuff2 := bytes.NewBufferString("")
	cmd2.SetOut(outBuff2)
	cmd2.SetArgs([]string{tmpFile.Name(), "KEY2"})
	err = cmd2.Execute()
	assert.NoError(t, err)
	out2, _ := io.ReadAll(outBuff2)
	assert.Equal(t, "multi\nline", string(out2))

	// Get KEY3
	cmd3 := newGetCmd()
	outBuff3 := bytes.NewBufferString("")
	cmd3.SetOut(outBuff3)
	cmd3.SetArgs([]string{tmpFile.Name(), "KEY3"})
	err = cmd3.Execute()
	assert.NoError(t, err)
	out3, _ := io.ReadAll(outBuff3)
	assert.Equal(t, "another", string(out3))
}

func TestMultiLineWithEscapedQuote(t *testing.T) {
	content := `KEY="line with \"quote
continues"
`
	tmpFile, err := createRandomTestFileWithContent(content)
	if err != nil {
		t.Fatalf("Error setting up test file: %v", err)
	}
	defer removeTestFile(tmpFile.Name())

	cmd := newGetCmd()
	outBuff := bytes.NewBufferString("")
	cmd.SetOut(outBuff)
	cmd.SetArgs([]string{tmpFile.Name(), "KEY"})

	err = cmd.Execute()
	assert.NoError(t, err)

	outContent, err := io.ReadAll(outBuff)
	assert.NoError(t, err)
	// The escaped quote backslash is preserved
	assert.Equal(t, "line with \\\"quote\ncontinues", string(outContent))
}
