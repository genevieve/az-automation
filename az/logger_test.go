package az_test

import (
	"bytes"

	"github.com/genevievelesperance/az-automation/az"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Logger", func() {
	var (
		buffer *bytes.Buffer
		logger *az.Logger
	)

	BeforeEach(func() {
		buffer = bytes.NewBuffer([]byte{})
		logger = az.NewLogger(buffer)
	})

	Describe("Println", func() {
		It("prints out the message", func() {
			logger.Println("banana")

			Expect(buffer.String()).To(Equal("banana\n"))
		})
	})
})
