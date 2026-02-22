package kvf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse_SimpleValue(t *testing.T) {
	parser := NewLineParser()
	item, err := parser.Parse("key=value")
	assert.NoError(t, err)
	assert.Equal(t, "key", item.Key)
	assert.Equal(t, "value", item.Val)
	assert.Equal(t, "", item.Quote)
	assert.Equal(t, ParseStateComplete, item.ParseState)
}

func TestParse_DoubleQuotedValue(t *testing.T) {
	parser := NewLineParser()
	item, err := parser.Parse(`key="value"`)
	assert.NoError(t, err)
	assert.Equal(t, "key", item.Key)
	assert.Equal(t, "value", item.Val)
	assert.Equal(t, `"`, item.Quote)
	assert.Equal(t, ParseStateComplete, item.ParseState)
}

func TestParse_SingleQuotedValue(t *testing.T) {
	parser := NewLineParser()
	item, err := parser.Parse(`key='value'`)
	assert.NoError(t, err)
	assert.Equal(t, "key", item.Key)
	assert.Equal(t, "value", item.Val)
	assert.Equal(t, `'`, item.Quote)
	assert.Equal(t, ParseStateComplete, item.ParseState)
}

func TestParse_MultiLineStart_DoubleQuote(t *testing.T) {
	parser := NewLineParser()
	item, err := parser.Parse(`key="line1`)
	assert.NoError(t, err)
	assert.Equal(t, "key", item.Key)
	assert.Equal(t, `"`, item.Quote)
	assert.Equal(t, ParseStateMultiLineStart, item.ParseState)
	assert.True(t, parser.IsInMultiLine())
	assert.Equal(t, "key", parser.GetCurrentKey())
}

func TestParse_MultiLineContinueAndComplete_DoubleQuote(t *testing.T) {
	parser := NewLineParser()

	// Start multi-line
	item1, err := parser.Parse(`key="line1`)
	assert.NoError(t, err)
	assert.Equal(t, ParseStateMultiLineStart, item1.ParseState)

	// Continue multi-line
	item2, err := parser.Parse("line2")
	assert.NoError(t, err)
	assert.Equal(t, ParseStateMultiLineContinue, item2.ParseState)
	assert.Equal(t, "line1\nline2", item2.Val)

	// Complete multi-line
	item3, err := parser.Parse(`line3"`)
	assert.NoError(t, err)
	assert.Equal(t, ParseStateComplete, item3.ParseState)
	assert.Equal(t, "line1\nline2\nline3", item3.Val)
	assert.False(t, parser.IsInMultiLine())
}

func TestParse_MultiLineStart_SingleQuote(t *testing.T) {
	parser := NewLineParser()
	item, err := parser.Parse(`key='line1`)
	assert.NoError(t, err)
	assert.Equal(t, "key", item.Key)
	assert.Equal(t, `'`, item.Quote)
	assert.Equal(t, ParseStateMultiLineStart, item.ParseState)
	assert.True(t, parser.IsInMultiLine())
}

func TestParse_MultiLineContinueAndComplete_SingleQuote(t *testing.T) {
	parser := NewLineParser()

	// Start multi-line
	item1, err := parser.Parse(`key='line1`)
	assert.NoError(t, err)
	assert.Equal(t, ParseStateMultiLineStart, item1.ParseState)

	// Continue multi-line
	item2, err := parser.Parse("line2")
	assert.NoError(t, err)
	assert.Equal(t, ParseStateMultiLineContinue, item2.ParseState)

	// Complete multi-line
	item3, err := parser.Parse(`line3'`)
	assert.NoError(t, err)
	assert.Equal(t, ParseStateComplete, item3.ParseState)
	assert.Equal(t, "line1\nline2\nline3", item3.Val)
	assert.False(t, parser.IsInMultiLine())
}

func TestParse_MultiLineWithEmptyLine(t *testing.T) {
	parser := NewLineParser()

	// Start multi-line
	_, err := parser.Parse(`key="line1`)
	assert.NoError(t, err)

	// Continue with content
	item2, _ := parser.Parse("line2")
	assert.Equal(t, "line1\nline2", item2.Val)

	// Empty line
	item3, _ := parser.Parse("")
	assert.Equal(t, "line1\nline2\n", item3.Val)

	// Complete
	item4, _ := parser.Parse(`line3"`)
	assert.Equal(t, "line1\nline2\n\nline3", item4.Val)
	assert.Equal(t, ParseStateComplete, item4.ParseState)
}

func TestParse_EscapedQuote_DoubleQuote(t *testing.T) {
	parser := NewLineParser()

	// Start multi-line with escaped quote in content
	item1, _ := parser.Parse(`key="line1 with \"quote`)
	assert.Equal(t, ParseStateMultiLineStart, item1.ParseState)

	// Complete - the escaped quote should not close the value
	item2, _ := parser.Parse(`continue"`)
	assert.Equal(t, ParseStateComplete, item2.ParseState)
	// The escaped quote backslash is preserved in the value
	assert.Equal(t, "line1 with \\\"quote\ncontinue", item2.Val)
}

func TestParse_EscapedQuote_SingleQuote(t *testing.T) {
	parser := NewLineParser()

	// Start multi-line with escaped quote in content
	item1, _ := parser.Parse(`key='line1 with \'quote`)
	assert.Equal(t, ParseStateMultiLineStart, item1.ParseState)

	// Complete - the escaped quote should not close the value
	item2, _ := parser.Parse(`continue'`)
	assert.Equal(t, ParseStateComplete, item2.ParseState)
	// The escaped quote backslash is preserved in the value
	assert.Equal(t, "line1 with \\'quote\ncontinue", item2.Val)
}

func TestParse_Comment(t *testing.T) {
	parser := NewLineParser()
	item, err := parser.Parse("# This is a comment")
	assert.NoError(t, err)
	assert.True(t, item.IsComment)
	assert.Equal(t, "This is a comment", item.Val)
	assert.Equal(t, ParseStateComplete, item.ParseState)
}

func TestParse_EmptyLine(t *testing.T) {
	parser := NewLineParser()
	item, err := parser.Parse("")
	assert.NoError(t, err)
	assert.True(t, item.IsEmpty)
	assert.Equal(t, ParseStateComplete, item.ParseState)
}

func TestParse_InvalidLine(t *testing.T) {
	parser := NewLineParser()
	_, err := parser.Parse("invalid_line_without_equals")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid line")
}

func TestParse_WithSpaces(t *testing.T) {
	parser := NewLineParser()
	item, err := parser.Parse("  key  =  value  ")
	assert.NoError(t, err)
	assert.Equal(t, "key", item.Key)
	assert.Equal(t, "value", item.Val)
}

func TestParse_MultiLineWithSpaces(t *testing.T) {
	parser := NewLineParser()

	item1, _ := parser.Parse(`  key  =  "line1`)
	assert.Equal(t, "key", item1.Key)
	assert.Equal(t, ParseStateMultiLineStart, item1.ParseState)

	item2, _ := parser.Parse(`line2"`)
	assert.Equal(t, ParseStateComplete, item2.ParseState)
	assert.Equal(t, "line1\nline2", item2.Val)
}

func TestParser_Reset(t *testing.T) {
	parser := NewLineParser()

	// Start multi-line
	_, err := parser.Parse(`key="line1`)
	assert.NoError(t, err)
	assert.True(t, parser.IsInMultiLine())

	// Reset
	parser.Reset()
	assert.False(t, parser.IsInMultiLine())
	assert.Equal(t, "", parser.GetCurrentKey())
}

func TestItem_ToLine_Simple(t *testing.T) {
	item := &Item{Key: "key", Val: "value", Quote: ""}
	assert.Equal(t, "key=value\n", item.ToLine())
}

func TestItem_ToLine_Quoted(t *testing.T) {
	item := &Item{Key: "key", Val: "value", Quote: `"`}
	assert.Equal(t, "key=\"value\"\n", item.ToLine())
}

func TestItem_ToLine_MultiLine(t *testing.T) {
	item := &Item{Key: "key", Val: "line1\nline2\nline3", Quote: `"`}
	expected := "key=\"line1\nline2\nline3\"\n"
	assert.Equal(t, expected, item.ToLine())
}

func TestItem_ToLine_Comment(t *testing.T) {
	item := &Item{IsComment: true, Val: "comment text"}
	assert.Equal(t, "# comment text\n", item.ToLine())
}

func TestItem_ToLine_Empty(t *testing.T) {
	item := &Item{IsEmpty: true}
	assert.Equal(t, "\n", item.ToLine())
}

func TestNewItem_WithNewlines(t *testing.T) {
	item, err := NewItem("key", "line1\nline2")
	assert.NoError(t, err)
	assert.Equal(t, "key", item.Key)
	assert.Equal(t, "line1\nline2", item.Val)
	assert.Equal(t, `"`, item.Quote) // Should auto-set double quotes for multi-line
}

func TestNewItem_EmptyKey(t *testing.T) {
	_, err := NewItem("", "value")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "key is empty")
}
