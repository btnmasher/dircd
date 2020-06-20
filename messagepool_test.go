package dircd_test

import (
	. "github.com/btnmasher/dircd"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MessagePool", func() {

	var (
		msgp *MessagePool
	)

	BeforeEach(func() {
		msgp = NewMessagePool(1)
	})

	Describe("gives a new message", func() {
		Context("when the queue is empty", func() {
			It("returns a newly allocated message", func() {
				Expect(len(msgp.Messages)).Should(Equal(0))
				msg := msgp.New()
				Expect(msg).ShouldNot(BeNil())
			})
		})
		Context("when the queue is not empty", func() {
			It("returns a new message from the queue", func() {
				msgp.Recycle(&Message{})
				Expect(len(msgp.Messages)).Should(Equal(1))
				msgp.New()
				Expect(len(msgp.Messages)).Should(Equal(0))
			})
		})
	})

	Describe("recycles a message", func() {
		It("should scrub the message of any state", func() {
			msg1 := &Message{
				Sender:  "irc.someserver.org",
				Code:    ReplyWelcome,
				Command: CmdPrivMsg,
				Params:  []string{"somenick"},
				Text:    "I am the server.",
			}

			msgp.Recycle(msg1)
			msg2 := msgp.New()
			Expect(msg2.Sender).Should(Equal(""))
			Expect(msg2.Code).Should(Equal(ReplyNone))
			Expect(msg2.Command).Should(Equal(""))
			Expect(msg2.Params).Should(BeNil())
			Expect(msg2.Text).Should(Equal(""))
		})

		Context("when the pool is not full", func() {
			It("accepts the message and stores it in the queue", func() {
				Expect(len(msgp.Messages)).Should(Equal(0))
				msgp.Recycle(&Message{})
				Expect(len(msgp.Messages)).Should(Equal(1))
			})
		})

		Context("when the pool is full", func() {
			It("accepts the message and unreferences it, leaving the queue size intact", func() {
				msg1 := &Message{}
				msgp.Recycle(msg1)
				Expect(len(msgp.Messages)).Should(Equal(1))
				msg2 := &Message{}
				msgp.Recycle(msg2)
				Expect(len(msgp.Messages)).Should(Equal(1))
			})
		})
	})
})
