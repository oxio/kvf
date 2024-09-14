package cmd

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

func TestGetValue_Success(t *testing.T) {
	content := `key1=value1
key2="value2"
# This is a comment
key3='value3'
`
	tmpFile, err := createRandomTestFileWithContent(content)
	if err == nil {
		defer removeTestFile(tmpFile.Name())
	}

	if err != nil {
		t.Fatalf("Error setting up test file: %v", err)
	}

	testCases := []struct {
		name          string
		key           string
		defaultValue  string
		expectedValue string
		expectedError string
	}{
		{
			name:          "Simple key",
			key:           "key1",
			expectedValue: "value1",
			expectedError: "",
		},
		{
			name:          "Key with double quotes",
			key:           "key2",
			expectedValue: "value2",
			expectedError: "",
		},
		{
			name:          "Key with single quotes",
			key:           "key3",
			expectedValue: "value3",
			expectedError: "",
		},
		{
			name:          "Not found key",
			key:           "bogus-key",
			expectedValue: "",
			expectedError: "key doesn't exist",
		},
		{
			name:          "Not found key with default",
			key:           "bogus-key",
			defaultValue:  "default-value",
			expectedValue: "default-value",
			expectedError: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := newGetCmd()
			outBuff := bytes.NewBufferString("")
			errBuff := bytes.NewBufferString("")
			cmd.SetOut(outBuff)
			cmd.SetErr(errBuff)
			if tc.defaultValue != "" {
				cmd.SetArgs([]string{tmpFile.Name(), tc.key, "--default", tc.defaultValue})
			} else {
				cmd.SetArgs([]string{tmpFile.Name(), tc.key})
			}

			err = cmd.Execute()
			if tc.expectedError != "" {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			outContent, err := io.ReadAll(outBuff)
			assert.NoError(t, err)
			errContent, err := io.ReadAll(errBuff)
			assert.NoError(t, err)

			if tc.expectedValue != "" {
				assert.Contains(t, string(outContent), tc.expectedValue)
			}
			if tc.expectedError != "" {
				assert.Contains(t, string(errContent), tc.expectedError)
			}
		})
	}
}

func TestGetValue_InvalidFileFormat(t *testing.T) {
	content := `key1=value1
key2
key3='value3'
`

	tmpFile, err := createRandomTestFileWithContent(content)
	if err == nil {
		defer removeTestFile(tmpFile.Name())
	}

	if err != nil {
		t.Fatalf("Error setting up test file: %v", err)
	}

	cmd := newGetCmd()
	outBuff := bytes.NewBufferString("")
	errBuff := bytes.NewBufferString("")
	cmd.SetOut(outBuff)
	cmd.SetErr(errBuff)
	cmd.SetArgs([]string{tmpFile.Name(), "foo"})
	err = cmd.Execute()

	assert.Error(t, err)
	errContent, err := io.ReadAll(errBuff)
	assert.Contains(t, string(errContent), "invalid line: key2")
}