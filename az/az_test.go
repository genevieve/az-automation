package az_test

import (
	"errors"
	"io/ioutil"
	"os"

	"github.com/genevievelesperance/az-automation/az"
	"github.com/genevievelesperance/az-automation/az/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Az", func() {
	var (
		azure *az.Az

		cli                  *fakes.CLI
		account              string
		displayName          string
		identifierUri        string
		credentialOutputFile string
	)

	BeforeEach(func() {
		cli = &fakes.CLI{}
		account = "some-account"
		displayName = "some-display-name"
		identifierUri = "http://some-identifier-uri"
		credentialOutputFile = "some-credential-file"

		azure = az.NewAz(cli, account, displayName, identifierUri, credentialOutputFile)
	})

	Context("ValidVersion", func() {
		BeforeEach(func() {
			cli.ExecuteCall.Returns.Output = "49.0.0"
		})

		It("checks the azure-cli is 2.0", func() {
			err := azure.ValidVersion()
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when this first execute call fails", func() {
			BeforeEach(func() {
				cli.ExecuteCall.Returns.Error = errors.New("some error")
			})

			It("returns a helpful error", func() {
				err := azure.ValidVersion()
				Expect(err).To(MatchError("Please install the azure-cli."))
			})
		})

		Context("when the cli version cannot be parsed", func() {
			BeforeEach(func() {
				cli.ExecuteCall.Returns.Output = "$.$.$"
			})

			It("returns a helpful error", func() {
				err := azure.ValidVersion()
				Expect(err).To(MatchError("The azure-cli version could not be parsed."))
			})
		})

		Context("when the cli version is below the minimum", func() {
			BeforeEach(func() {
				cli.ExecuteCall.Returns.Output = "1.0.0"
			})

			It("returns a helpful error", func() {
				err := azure.ValidVersion()
				Expect(err).To(MatchError("Please update the azure-cli to at least 2.0.0."))
			})
		})
	})

	PContext("LoggedIn", func() {})

	PContext("GetSubscriptionAndTenantId", func() {})

	PContext("AppExists", func() {})

	PContext("CreateApplication", func() {})

	PContext("CreateServicePrincipal", func() {})

	PContext("AssignContributorRole", func() {})

	Context("WriteCredentials", func() {
		AfterEach(func() {
			err := os.Remove("some-credential-file")
			Expect(err).NotTo(HaveOccurred())
		})

		It("writes the credentials to the specified output file", func() {
			err := azure.WriteCredentials()
			Expect(err).NotTo(HaveOccurred())

			bytes, err := ioutil.ReadFile(credentialOutputFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(bytes)).To(ContainSubstring("subscription_id ="))
			Expect(string(bytes)).To(ContainSubstring("tenant_id ="))
			Expect(string(bytes)).To(ContainSubstring("client_id ="))
			Expect(string(bytes)).To(ContainSubstring("client_secret ="))
		})
	})
})
