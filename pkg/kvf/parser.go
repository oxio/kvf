package kvf

import (
	"fmt"
	"strings"
)

// unescapeValue unescapes backslashes and quote characters in a value
func unescapeValue(val string, quote string) string {
	var result strings.Builder
	escaped := false

	for _, ch := range val {
		if escaped {
			if string(ch) == quote {
				result.WriteString(quote)
			} else if ch == '\\' {
				result.WriteRune('\\')
			} else {
				// Preserve unknown escape sequences as-is
				result.WriteRune('\\')
				result.WriteRune(ch)
			}
			escaped = false
		} else if ch == '\\' {
			escaped = true
		} else {
			result.WriteRune(ch)
		}
	}
	// Handle trailing backslash
	if escaped {
		result.WriteRune('\\')
	}

	return result.String()
}

type LineParser struct {
	inMultiLine  bool
	currentKey   string
	currentQuote string
	lineBuffer   strings.Builder
}

func NewLineParser() *LineParser {
	return &LineParser{}
}

// Reset clears the parser state for reuse
func (p *LineParser) Reset() {
	p.inMultiLine = false
	p.currentKey = ""
	p.currentQuote = ""
	p.lineBuffer.Reset()
}

// Parse parses a single line and returns an Item.
// For multi-line values, it returns items with ParseStateMultiLineStart/Continue
// until the closing quote is found.
func (p *LineParser) Parse(line string) (*Item, error) {
	// If we're in the middle of a multi-line value, continue accumulating
	if p.inMultiLine {
		return p.continueMultiLine(line)
	}

	line = strings.TrimSpace(line)

	item := &Item{ParseState: ParseStateComplete}

	if line == "" {
		return &Item{IsEmpty: true, ParseState: ParseStateComplete}, nil
	} else if strings.HasPrefix(line, "#") {
		item.IsComment = true
		item.Val = strings.TrimSpace(strings.TrimLeft(line, "#"))
	} else {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid line: %s", line)
		}
		item.Key = strings.TrimSpace(parts[0])
		item.Val = strings.TrimSpace(parts[1])

		// Check for quoted values
		if strings.HasPrefix(item.Val, "\"") {
			item.Quote = "\""
		} else if strings.HasPrefix(item.Val, "'") {
			item.Quote = "'"
		}

		// Handle quoted values
		if item.Quote != "" {
			// Check if this is a multi-line value (quote not closed on same line)
			if !p.isCompleteQuotedValue(item.Val, item.Quote) {
				// Start multi-line mode
				p.inMultiLine = true
				p.currentKey = item.Key
				p.currentQuote = item.Quote
				p.lineBuffer.Reset()
				// Store the value without the opening quote
				p.lineBuffer.WriteString(item.Val[1:])
				item.ParseState = ParseStateMultiLineStart
				return item, nil
			}
			// Single-line quoted value - strip quotes and unescape
			item.Val = unescapeValue(item.Val[1:len(item.Val)-1], item.Quote)
		}
	}

	return item, nil
}

// continueMultiLine handles lines that are part of a multi-line value
func (p *LineParser) continueMultiLine(line string) (*Item, error) {
	item := &Item{
		Key:        p.currentKey,
		Quote:      p.currentQuote,
		ParseState: ParseStateMultiLineContinue,
	}

	// Check if this line contains the closing quote
	closingIdx := p.findClosingQuote(line, p.currentQuote)

	if closingIdx >= 0 {
		// Found closing quote - complete the multi-line value
		p.lineBuffer.WriteString("\n")
		p.lineBuffer.WriteString(line[:closingIdx])
		item.Val = unescapeValue(p.lineBuffer.String(), p.currentQuote)
		item.ParseState = ParseStateComplete

		// Reset parser state
		p.Reset()
	} else {
		// Still continuing - add newline and append
		if p.lineBuffer.Len() > 0 {
			p.lineBuffer.WriteString("\n")
		}
		p.lineBuffer.WriteString(line)
		item.Val = p.lineBuffer.String()
	}

	return item, nil
}

// isCompleteQuotedValue checks if a quoted value is complete on a single line
// It must start with the quote, end with the quote, and the ending quote must not be escaped
func (p *LineParser) isCompleteQuotedValue(val string, quote string) bool {
	if len(val) < 2 {
		return false
	}
	// Must start with the quote
	if !strings.HasPrefix(val, quote) {
		return false
	}
	// Find the closing quote (skipping the opening quote)
	closingIdx := p.findClosingQuote(val[1:], quote)
	if closingIdx < 0 {
		return false
	}
	// The closing quote should be at the end of the value
	return closingIdx == len(val)-2 // -2 because we skipped the first character
}

// findClosingQuote finds the index of the closing quote in a line,
// respecting escaped quotes. Returns -1 if not found.
func (p *LineParser) findClosingQuote(line string, quote string) int {
	escapeChar := `\`
	escaped := false

	for i, ch := range line {
		if escaped {
			escaped = false
			continue
		}
		if string(ch) == escapeChar {
			escaped = true
			continue
		}
		if string(ch) == quote {
			return i
		}
	}
	return -1
}

// IsInMultiLine returns true if the parser is currently in multi-line mode
func (p *LineParser) IsInMultiLine() bool {
	return p.inMultiLine
}

// GetCurrentKey returns the key being parsed in multi-line mode
func (p *LineParser) GetCurrentKey() string {
	return p.currentKey
}
