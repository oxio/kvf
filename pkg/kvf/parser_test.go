package kvf

import (
	"strings"
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
	// The escaped quote is now unescaped when reading
	assert.Equal(t, "line1 with \"quote\ncontinue", item2.Val)
}

func TestParse_EscapedQuote_SingleQuote(t *testing.T) {
	parser := NewLineParser()

	// Start multi-line with escaped quote in content
	item1, _ := parser.Parse(`key='line1 with \'quote`)
	assert.Equal(t, ParseStateMultiLineStart, item1.ParseState)

	// Complete - the escaped quote should not close the value
	item2, _ := parser.Parse(`continue'`)
	assert.Equal(t, ParseStateComplete, item2.ParseState)
	// The escaped quote is now unescaped when reading
	assert.Equal(t, "line1 with 'quote\ncontinue", item2.Val)
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

// =============================================================================
// Quote Escaping Tests
// =============================================================================

func TestNewItem_WithDoubleQuotes(t *testing.T) {
	item, err := NewItem("key", `he said "hello"`)
	assert.NoError(t, err)
	assert.Equal(t, "key", item.Key)
	assert.Equal(t, `he said "hello"`, item.Val)
	assert.Equal(t, `'`, item.Quote) // Should use single quotes to wrap double quotes
}

func TestNewItem_WithSingleQuotes(t *testing.T) {
	item, err := NewItem("key", `it's mine`)
	assert.NoError(t, err)
	assert.Equal(t, "key", item.Key)
	assert.Equal(t, `it's mine`, item.Val)
	assert.Equal(t, `"`, item.Quote) // Should use double quotes to wrap single quotes
}

func TestNewItem_WithBothQuotes(t *testing.T) {
	item, err := NewItem("key", `both " and '`)
	assert.NoError(t, err)
	assert.NoError(t, err)
	assert.Equal(t, "key", item.Key)
	assert.Equal(t, `both " and '`, item.Val)
	assert.Equal(t, `"`, item.Quote) // Must use double quotes, will need escaping
}

func TestNewItem_WithBackslash(t *testing.T) {
	item, err := NewItem("key", `path\to\file`)
	assert.NoError(t, err)
	assert.Equal(t, "key", item.Key)
	assert.Equal(t, `path\to\file`, item.Val)
	assert.Equal(t, `"`, item.Quote) // Should use quotes for backslash
}

func TestNewItem_NoQuotesNeeded(t *testing.T) {
	item, err := NewItem("key", "simple value")
	assert.NoError(t, err)
	assert.Equal(t, "key", item.Key)
	assert.Equal(t, "simple value", item.Val)
	assert.Equal(t, "", item.Quote) // No quotes needed
}

func TestToLine_WithDoubleQuotes_Escaped(t *testing.T) {
	item := &Item{Key: "key", Val: `he said "hello"`, Quote: `"`}
	// Double quotes should be escaped
	assert.Equal(t, `key="he said \"hello\""`+"\n", item.ToLine())
}

func TestToLine_WithSingleQuotes_Wrapped(t *testing.T) {
	item := &Item{Key: "key", Val: `he said "hello"`, Quote: `'`}
	// Single quotes wrapper - no escaping needed
	assert.Equal(t, `key='he said "hello"'`+"\n", item.ToLine())
}

func TestToLine_WithBackslash_Escaped(t *testing.T) {
	item := &Item{Key: "key", Val: `path\to\file`, Quote: `"`}
	// Backslashes should be escaped
	assert.Equal(t, `key="path\\to\\file"`+"\n", item.ToLine())
}

func TestToLine_WithBothQuotes_Escaped(t *testing.T) {
	item := &Item{Key: "key", Val: `both " and '`, Quote: `"`}
	// Double quotes should be escaped, single quotes preserved
	assert.Equal(t, `key="both \" and '"`+"\n", item.ToLine())
}

func TestParse_SingleLine_WithEscapedQuotes(t *testing.T) {
	parser := NewLineParser()
	item, err := parser.Parse(`key="value with \"quote\""`)
	assert.NoError(t, err)
	assert.Equal(t, "key", item.Key)
	assert.Equal(t, `value with "quote"`, item.Val)
	assert.Equal(t, ParseStateComplete, item.ParseState)
}

func TestParse_SingleLine_WithEscapedBackslash(t *testing.T) {
	parser := NewLineParser()
	item, err := parser.Parse(`key="path\\to\\file"`)
	assert.NoError(t, err)
	assert.Equal(t, "key", item.Key)
	assert.Equal(t, `path\to\file`, item.Val)
	assert.Equal(t, ParseStateComplete, item.ParseState)
}

func TestRoundTrip_ValueWithQuotes(t *testing.T) {
	originalValue := `he said "hello"`

	// Create item (writing)
	item, err := NewItem("key", originalValue)
	assert.NoError(t, err)

	// Convert to line
	line := item.ToLine()

	// Parse back (reading)
	parser := NewLineParser()
	parsedItem, err := parser.Parse(strings.TrimSpace(line))
	assert.NoError(t, err)

	// Should get original value back
	assert.Equal(t, originalValue, parsedItem.Val)
}

func TestRoundTrip_ValueWithBackslash(t *testing.T) {
	originalValue := `path\to\file`

	// Create item (writing)
	item, err := NewItem("key", originalValue)
	assert.NoError(t, err)

	// Convert to line
	line := item.ToLine()

	// Parse back (reading)
	parser := NewLineParser()
	parsedItem, err := parser.Parse(strings.TrimSpace(line))
	assert.NoError(t, err)

	// Should get original value back
	assert.Equal(t, originalValue, parsedItem.Val)
}

func TestRoundTrip_ValueWithBothQuotes(t *testing.T) {
	originalValue := `both " and ' quotes`

	// Create item (writing)
	item, err := NewItem("key", originalValue)
	assert.NoError(t, err)

	// Convert to line
	line := item.ToLine()

	// Parse back (reading)
	parser := NewLineParser()
	parsedItem, err := parser.Parse(strings.TrimSpace(line))
	assert.NoError(t, err)

	// Should get original value back
	assert.Equal(t, originalValue, parsedItem.Val)
}

func TestRoundTrip_MultiLineWithQuotes(t *testing.T) {
	originalValue := "line1 with \"quote\"\nline2"

	// Create item (writing)
	item, err := NewItem("key", originalValue)
	assert.NoError(t, err)
	assert.Equal(t, `"`, item.Quote)

	// Convert to line
	line := item.ToLine()

	// Parse back (reading) - multi-line needs two parse calls
	parser := NewLineParser()
	parsedItem1, err := parser.Parse(strings.Split(line, "\n")[0])
	assert.NoError(t, err)
	assert.Equal(t, ParseStateMultiLineStart, parsedItem1.ParseState)

	parsedItem2, err := parser.Parse(strings.Split(line, "\n")[1])
	assert.NoError(t, err)
	assert.Equal(t, ParseStateComplete, parsedItem2.ParseState)

	// Should get original value back
	assert.Equal(t, originalValue, parsedItem2.Val)
}
