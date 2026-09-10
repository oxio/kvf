package cmd

import (
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"testing"
)

// Runs the root command the same way main() does: with the output writer set
// to os.Stdout and both standard streams captured.
func executeRootCmdCaptured(t *testing.T, args ...string) (string, string, error) {
	t.Helper()

	origStdout, origStderr := os.Stdout, os.Stderr
	stdoutR, stdoutW, errPipe := os.Pipe()
	assert.NoError(t, errPipe)
	stderrR, stderrW, errPipe2 := os.Pipe()
	assert.NoError(t, errPipe2)
	os.Stdout, os.Stderr = stdoutW, stderrW

	t.Cleanup(func() {
		os.Stdout, os.Stderr = origStdout, origStderr
		_ = stdoutW.Close()
		_ = stderrW.Close()
		_ = stdoutR.Close()
		_ = stderrR.Close()
		SetOut(origStdout)
		rootCmd.SetArgs(nil)
	})

	SetOut(os.Stdout)
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()

	_ = stdoutW.Close()
	_ = stderrW.Close()
	outBytes, err := io.ReadAll(stdoutR)
	assert.NoError(t, err)
	errBytes, err := io.ReadAll(stderrR)
	assert.NoError(t, err)
	return string(outBytes), string(errBytes), nil
}

// Verifies that "get" writes the value to stdout (not stderr) when executed
// through the root command with the output writer set by main().
func TestGetOutputIsWrittenToStdout(t *testing.T) {
	tmpFile, err := createRandomTestFileWithContent("key=value1")
	assert.NoError(t, err)
	defer removeTestFile(tmpFile.Name())

	stdout, stderr, err := executeRootCmdCaptured(t, "get", tmpFile.Name(), "key")
	assert.NoError(t, err)

	assert.Equal(t, "value1", stdout)
	assert.Empty(t, stderr)
}

// Verifies that "set" writes the value to the file and produces no output
// on either stdout or stderr.
func TestSetWritesValueToFilesWithNoOutput(t *testing.T) {
	filePath := getRandomTestFilePath()

	stdout, stderr, err := executeRootCmdCaptured(t, "set", filePath, "key", "value1")
	if err == nil {
		defer removeTestFile(filePath)
	}
	assert.NoError(t, err)

	assertFileContentEquals(t, filePath, "key=value1\n")
	assert.Empty(t, stdout)
	assert.Empty(t, stderr)
}
