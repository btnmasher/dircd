package dircd_test

import (
	"encoding/json"

	. "github.com/btnmasher/dircd"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Message", func() {
	var (
		validMsg   Message
		welcomeMsg Message
	)

	BeforeEach(func() {
		validMsg = Message{
			Sender:  "irc.someserver.net",
			Command: CmdPrivMsg,
			Params:  []string{"nick1!someuser@irc.somehost.org"},
			Text:    "I am the server",
		}
		welcomeMsg = Message{
			Sender: "irc.someserver.net",
			Code:   ReplyWelcome,
			Params: []string{"nick1!someuser@irc.somehost.org"},
			Text:   "Welcome to the server",
		}
	})

	Describe("renders to string", func() {
		Context("when the command is not a numeric code", func() {
			It("should be valid IRC message", func() {
				str := validMsg.Render()
				Expect(str).Should(Equal(":irc.someserver.net PRIVMSG nick1!someuser@irc.somehost.org :I am the server\r\n"))
			})
		})

		Context("when the command is a numeric code", func() {
			It("should be a valid IRC message", func() {
				str := welcomeMsg.Render()
				Expect(str).Should(Equal(":irc.someserver.net 001 nick1!someuser@irc.somehost.org :Welcome to the server\r\n"))
			})
		})

		Context("when using the stringer interface function with a valid message", func() {
			It("should be a valid IRC message", func() {
				Expect(welcomeMsg.String()).Should(Equal(":irc.someserver.net 001 nick1!someuser@irc.somehost.org :Welcome to the server\r\n"))
			})
		})
	})

	Describe("renders to debug JSON", func() {
		Context("given a valid Message", func() {
			It("should return valid JSON", func() {
				str := validMsg.Debug()
				var msg Message
				err := json.Unmarshal([]byte(str), &msg)
				Expect(err).ShouldNot(HaveOccurred())
			})
		})
	})

})
