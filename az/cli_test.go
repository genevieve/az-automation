package az_test

import (
	"os/exec"

	"github.com/genevievelesperance/az-automation/az"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CLI", func() {
	var cli az.CLI

	BeforeEach(func() {
		path, err := exec.LookPath("echo")
		if err != nil {
			Skip("Failed to locate echo.")
		}

		cli = az.NewCLI(path)
	})

	It("returns the output of the command it executed", func() {
		output, err := cli.Execute([]string{"fake", "arg"})
		Expect(err).NotTo(HaveOccurred())

		Expect(output).To(ContainSubstring("fake arg"))
	})
})
