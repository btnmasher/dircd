package dircd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessage(t *testing.T) {
	tests := []struct {
		name     string
		msg      Message
		expected string
	}{
		{
			name: "valid message",
			msg: Message{
				Source:   "irc.someserver.net",
				Command:  CmdPrivMsg,
				Params:   []string{"nick1!someuser@irc.somehost.org"},
				Trailing: "I am the server",
			},
			expected: ":irc.someserver.net PRIVMSG nick1!someuser@irc.somehost.org :I am the server\r\n",
		},
		{
			name: "numeric code message",
			msg: Message{
				Source:   "irc.someserver.net",
				Code:     ReplyWelcome,
				Params:   []string{"nick1!someuser@irc.somehost.org"},
				Trailing: "Welcome to the server",
			},
			expected: ":irc.someserver.net 001 nick1!someuser@irc.somehost.org :Welcome to the server\r\n",
		},
		{
			name: "stringer interface function",
			msg: Message{
				Source:   "irc.someserver.net",
				Code:     ReplyWelcome,
				Params:   []string{"nick1!someuser@irc.somehost.org"},
				Trailing: "Welcome to the server",
			},
			expected: ":irc.someserver.net 001 nick1!someuser@irc.somehost.org :Welcome to the server\r\n",
		},
		{
			name: "debug function",
			msg: Message{
				Source:   "irc.someserver.net",
				Code:     ReplyWelcome,
				Params:   []string{"nick1!someuser@irc.somehost.org"},
				Trailing: "Welcome to the server",
			},
			expected: `{"sender":"irc.someserver.net","code":1,"params":["nick1!someuser@irc.somehost.org"],"text":"Welcome to the server","command":""}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.name {
			case "valid message", "numeric code message":
				assert.Equal(t, tt.expected, tt.msg.Render())
			case "stringer interface function":
				assert.Equal(t, tt.expected, tt.msg.String())
			case "debug function":
				assert.JSONEq(t, tt.expected, tt.msg.Debug())
			}
		})
	}
}
