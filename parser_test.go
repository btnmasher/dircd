package dircd_test

import (
	"fmt"
	"strings"

	. "github.com/btnmasher/dircd"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Parser", func() {
	var (
		validMsg       string = "PRIVMSG nick1!someuser@irc.somehost.org :I am the client\r\n" // noCRLF string = "PRIVMSG nick1!someuser@irc.somehost.org :I am the client"
		tooManyParams  string = "PRIVMSG 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 :I am the client\r\n"
		clientPrefixed string = ":prefix PRIVMSG nick1!someuser@irc.somehost.org :I am the client\r\n"
		tooSmall       string = "abc"
		tooLong        string = fmt.Sprint(strings.Repeat("a", MaxMsgLength), "\r\n")
		allWhitespace  string = "   \r\n"
	)

	Describe("given a valid string", func() {
		It("Should not return an error", func() {
			_, err := Parse(validMsg)
			Expect(err).ShouldNot(HaveOccurred())
		})

		Context("with a PRIVMSG", func() {
			msg, err := Parse(validMsg)

			It("should not return an error", func() {
				Expect(err).ShouldNot(HaveOccurred())
			})

			It("should have a valid command, a parameter, and a message text", func() {
				Expect(msg.Command).Should(Equal(CmdPrivMsg))
				Expect(msg.Params[0]).Should(Equal("nick1!someuser@irc.somehost.org"))
				Expect(msg.Text).Should(Equal("I am the client"))
			})
		})
	})

	Describe("given an invalid string", func() {
		Context("that is too short", func() {
			It("should return an error", func() {
				_, err := Parse(tooSmall)
				Expect(err).Should(Equal(ErrNotEnoughData))
			})
		})

		Context("that is too long", func() {
			It("should return an error", func() {
				_, err := Parse(tooLong)
				Expect(err).Should(Equal(ErrDataTooLong))
			})
		})

		// Context("with no CRLF", func() {
		// 	It("should return an error", func() {
		// 		_, err := Parse(noCRLF)
		// 		Expect(err).Should(Equal(ErrCRLF))
		// 	})
		// })

		Context("that is all whitespace", func() {
			It("should return an error", func() {
				_, err := Parse(allWhitespace)
				Expect(err).Should(Equal(ErrWhitespace))
			})
		})

		Context("with a prefix sent from the client", func() {
			It("should return an error", func() {
				_, err := Parse(clientPrefixed)
				Expect(err).Should(Equal(ErrPrefixed))
			})
		})

		Context("with too many parameters", func() {
			It("should return an error", func() {
				_, err := Parse(tooManyParams)
				Expect(err).Should(Equal(ErrTooManyParams))
			})
		})
	})
})
