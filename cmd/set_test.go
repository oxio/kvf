package cmd

import (
	"bytes"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"strconv"
	"sync"
	"testing"
)

func TestSetValue_Success(t *testing.T) {
	filePath := getRandomTestFilePath()
	cmd, _, _ := setUpTestSetCmd()
	cmd.SetArgs([]string{filePath, "key2", "value2"})

	err := cmd.Execute()
	if err == nil {
		defer removeTestFile(filePath)
	}
	assert.NoError(t, err)

	assertFileContentEquals(t, filePath, "key2=value2\n")
}

func TestOverrideValue_Success(t *testing.T) {
	filePath := getRandomTestFilePath()
	defer removeTestFile(filePath)

	cmd1, _, _ := setUpTestSetCmd()
	cmd1.SetArgs([]string{filePath, "key", "first value"})

	cmd2, _, _ := setUpTestSetCmd()
	cmd2.SetArgs([]string{filePath, "key", "second value"})

	err := cmd1.Execute()
	assert.NoError(t, err)

	assertFileContentEquals(t, filePath, "key=first value\n")

	err = cmd2.Execute()
	assert.NoError(t, err)

	assertFileContentEquals(t, filePath, "key=second value\n")
}

func TestExistingFileWithEmptyLinesAndComments(t *testing.T) {
	initialContent := `key1=value1

# Comment within empty lines

key2='value2'

`

	expectedContentAfterUpdate := `key1=value1

# Comment within empty lines

key2='value333'

`

	testFile, err := createRandomTestFileWithContent(initialContent)
	if err == nil {
		defer removeTestFile(testFile.Name())
	}

	if err != nil {
		t.Fatalf("Error setting up test file: %v", err)
	}

	cmd, _, _ := setUpTestSetCmd()
	cmd.SetArgs([]string{testFile.Name(), "key2", "value333"})
	err = cmd.Execute()
	assert.NoError(t, err)

	assertFileContentEquals(t, testFile.Name(), expectedContentAfterUpdate)
}

func TestSetMultipleFiles_Success(t *testing.T) {
	filePath1 := getRandomTestFilePath()
	filePath2 := getRandomTestFilePath()
	cmd, _, _ := setUpTestSetCmd()
	cmd.SetArgs([]string{filePath1, filePath2, "foo", "bar"})

	err := cmd.Execute()
	if err == nil {
		defer removeTestFile(filePath1)
		defer removeTestFile(filePath2)
	}
	assert.NoError(t, err)

	assertFileContentEquals(t, filePath1, "foo=bar\n")
	assertFileContentEquals(t, filePath2, "foo=bar\n")
}

func TestMultipleConcurrentWrites_Success(t *testing.T) {
	filePath := getRandomTestFilePath()
	defer removeTestFile(filePath)

	var wg sync.WaitGroup

	var commandIds []string
	for i := 0; i < 4000; i++ {
		cmd, _, _ := setUpTestSetCmd()
		commandId := strconv.Itoa(i)
		commandIds = append(commandIds, commandId)
		cmd.SetArgs([]string{filePath, "key" + commandId, "value" + commandId})
		wg.Add(1)
		go func(cmd *cobra.Command) {
			defer wg.Done()
			err := cmd.Execute()
			assert.NoError(t, err)
		}(cmd)
	}

	wg.Wait()

	for _, commandId := range commandIds {
		expected := "key" + commandId + "=value" + commandId
		assertFileContentContains(t, filePath, expected)
	}
}

func TestMultipleSetsOfTheSameKeyCausingNewLinesToBeAdded_Regression(t *testing.T) {
	testFile, err := createRandomTestFileWithContent("key=\"foo\"\n")
	assert.NoError(t, err)
	defer removeTestFile(testFile.Name())

	for i := 0; i <= 3; i++ {
		cmd, _, _ := setUpTestSetCmd()
		cmd.SetArgs([]string{testFile.Name(), "key", "value with spaces " + strconv.Itoa(i)})

		err = cmd.Execute()
		assert.NoError(t, err)
	}

	assertFileContentEquals(t, testFile.Name(), "key=\"value with spaces 3\"\n")
}

func TestMultipleSetsOfTheSameKeyCausingFileBeingCorrupted_Regression(t *testing.T) {
	testFile, err := createRandomTestFileWithContent("key=\"foo bar\"\n")
	assert.NoError(t, err)
	defer removeTestFile(testFile.Name())

	cmd, _, _ := setUpTestSetCmd()
	cmd.SetArgs([]string{testFile.Name(), "key", "foo"})

	err = cmd.Execute()
	assert.NoError(t, err)

	assertFileContentEquals(t, testFile.Name(), "key=\"foo\"\n")
}

func TestPreserveNewLinesCount(t *testing.T) {
	testCases := []struct {
		name            string
		originalContent string
		key             string
		newValue        string
		expectedContent string
	}{
		{
			name:            "single starting line line - will always add empty line at end",
			originalContent: "key=\"foo bar\"",
			key:             "key",
			newValue:        "foo",
			expectedContent: "key=\"foo\"\n",
		},
		{
			name:            "single line with empty at end",
			originalContent: "key=\"foo bar\"\n",
			key:             "key",
			newValue:        "foo",
			expectedContent: "key=\"foo\"\n",
		},
		{
			name:            "two empty ending lines",
			originalContent: "key=\"foo bar\"\n\n",
			key:             "key",
			newValue:        "foo",
			expectedContent: "key=\"foo\"\n\n",
		},
		{
			name:            "three empty ending lines",
			originalContent: "key=\"foo bar\"\n\n\n",
			key:             "key",
			newValue:        "foo",
			expectedContent: "key=\"foo\"\n\n\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testFile, err := createRandomTestFileWithContent(tc.originalContent)
			assert.NoError(t, err)
			defer removeTestFile(testFile.Name())

			cmd, _, _ := setUpTestSetCmd()
			cmd.SetArgs([]string{testFile.Name(), tc.key, tc.newValue})

			err = cmd.Execute()
			assert.NoError(t, err)

			assertFileContentEquals(t, testFile.Name(), tc.expectedContent)
		})
	}
}

func setUpTestSetCmd() (*cobra.Command, *bytes.Buffer, *bytes.Buffer) {
	cmd := newSetCmd()
	outBuff := bytes.NewBufferString("")
	errBuff := bytes.NewBufferString("")
	cmd.SetOut(outBuff)
	cmd.SetErr(errBuff)

	return cmd, outBuff, errBuff
}

// =============================================================================
// Quote Escaping Integration Tests
// =============================================================================

func TestSetValue_WithDoubleQuotes(t *testing.T) {
	filePath := getRandomTestFilePath()
	cmd, _, _ := setUpTestSetCmd()
	cmd.SetArgs([]string{filePath, "message", `he said "hello"`})

	err := cmd.Execute()
	if err == nil {
		defer removeTestFile(filePath)
	}
	assert.NoError(t, err)

	// Should use single quotes to wrap double quotes (no escaping needed)
	assertFileContentEquals(t, filePath, `message='he said "hello"'`+"\n")
}

func TestSetValue_WithSingleQuotes(t *testing.T) {
	filePath := getRandomTestFilePath()
	cmd, _, _ := setUpTestSetCmd()
	cmd.SetArgs([]string{filePath, "message", `it's mine`})

	err := cmd.Execute()
	if err == nil {
		defer removeTestFile(filePath)
	}
	assert.NoError(t, err)

	// Should use double quotes to wrap single quotes (no escaping needed)
	assertFileContentEquals(t, filePath, `message="it's mine"`+"\n")
}

func TestSetValue_WithBothQuotes(t *testing.T) {
	filePath := getRandomTestFilePath()
	cmd, _, _ := setUpTestSetCmd()
	cmd.SetArgs([]string{filePath, "message", `both " and ' quotes`})

	err := cmd.Execute()
	if err == nil {
		defer removeTestFile(filePath)
	}
	assert.NoError(t, err)

	// Must use double quotes and escape the double quotes inside
	assertFileContentEquals(t, filePath, `message="both \" and ' quotes"`+"\n")
}

func TestSetValue_WithBackslash(t *testing.T) {
	filePath := getRandomTestFilePath()
	cmd, _, _ := setUpTestSetCmd()
	cmd.SetArgs([]string{filePath, "path", `C:\Users\test`})

	err := cmd.Execute()
	if err == nil {
		defer removeTestFile(filePath)
	}
	assert.NoError(t, err)

	// Should escape backslashes
	assertFileContentEquals(t, filePath, `path="C:\\Users\\test"`+"\n")
}

func TestSetValue_RoundTrip_WithQuotes(t *testing.T) {
	filePath := getRandomTestFilePath()
	originalValue := `he said "hello" and she said 'hi'`

	// Set the value
	cmdSet, _, _ := setUpTestSetCmd()
	cmdSet.SetArgs([]string{filePath, "message", originalValue})
	err := cmdSet.Execute()
	if err == nil {
		defer removeTestFile(filePath)
	}
	assert.NoError(t, err)

	// Get the value back
	cmdGet := newGetCmd()
	outBuff := bytes.NewBufferString("")
	cmdGet.SetOut(outBuff)
	cmdGet.SetArgs([]string{filePath, "message"})
	err = cmdGet.Execute()
	assert.NoError(t, err)

	// Should get the original value back
	assert.Equal(t, originalValue, outBuff.String())
}
