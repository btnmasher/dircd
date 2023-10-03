package dircd

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParser(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected error
	}{
		{
			name:     "valid message",
			input:    "PRIVMSG nick1!someuser@irc.somehost.org :I am the client\r\n",
			expected: nil,
		},
		{
			name:     "too many parameters",
			input:    "PRIVMSG 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 :I am the client\r\n",
			expected: ErrTooManyParams,
		},
		{
			name:     "client prefixed",
			input:    ":prefix PRIVMSG nick1!someuser@irc.somehost.org :I am the client\r\n",
			expected: ErrPrefixed,
		},
		{
			name:     "too small",
			input:    "abc",
			expected: ErrMessageTooShort,
		},
		{
			name:     "too long",
			input:    fmt.Sprint(strings.Repeat("a", MaxMsgLength), "\r\n"),
			expected: ErrMessageTooLong,
		},
		{
			name:     "all whitespace",
			input:    "   \r\n",
			expected: ErrWhitespace,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.input)
			assert.Equal(t, tt.expected, err)
		})
	}
}
