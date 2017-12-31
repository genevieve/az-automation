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
		logger               *fakes.Logger
		account              string
		displayName          string
		identifierUri        string
		credentialOutputFile string
	)

	BeforeEach(func() {
		cli = &fakes.CLI{}
		logger = &fakes.Logger{}
		account = "some-account"
		displayName = "some-display-name"
		identifierUri = "http://some-identifier-uri"
		credentialOutputFile = "some-credential-file"

		azure = az.NewAz(cli, logger, account, displayName, identifierUri, credentialOutputFile)
	})

	Describe("ValidVersion", func() {
		BeforeEach(func() {
			cli.ExecuteCall.Returns.Output = "49.0.0"
		})

		It("checks the azure-cli is 2.0", func() {
			err := azure.ValidVersion()
			Expect(err).NotTo(HaveOccurred())

			Expect(cli.ExecuteCall.Receives.Args).To(Equal([]string{"-v"}))
			Expect(logger.PrintlnCall.Receives.Message).To(Equal("Checked version of azure-cli is above 2.0.0."))
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

	Describe("LoggedIn", func() {
		BeforeEach(func() {
			cli.ExecuteCall.Returns.Output = `{"name": "some-account", "id": "some-id", "tenantId": "some-tenant-id"}`
		})

		It("checks the user is logged in", func() {
			account, err := azure.LoggedIn()
			Expect(err).NotTo(HaveOccurred())

			Expect(cli.ExecuteCall.Receives.Args).To(Equal([]string{"account", "show", "-s", "some-account"}))
			Expect(account.Name).To(Equal("some-account"))
			Expect(account.Id).To(Equal("some-id"))
			Expect(account.TenantId).To(Equal("some-tenant-id"))
			Expect(logger.PrintlnCall.Receives.Message).To(Equal("Checked you are logged in to the azure-cli."))
		})

		Context("when the cli returns an error", func() {
			BeforeEach(func() {
				cli.ExecuteCall.Returns.Error = errors.New("some error")
			})

			It("checks the user is logged in", func() {
				_, err := azure.LoggedIn()
				Expect(err).To(MatchError("Please login to the azure-cli."))
			})
		})

		Context("when the account json is invalid", func() {
			BeforeEach(func() {
				cli.ExecuteCall.Returns.Output = `{$$$}`
			})

			It("returns a helpful error", func() {
				_, err := azure.LoggedIn()
				Expect(err).To(MatchError(ContainSubstring("Unmarshalling account json: ")))
			})
		})
	})

	Describe("GetSubscriptionAndTenantId", func() {
		var account az.Account
		BeforeEach(func() {
			account.Id = "some-id"
			account.TenantId = "some-tenant-id"
		})

		It("returns the subscription and tenant id", func() {
			id, tenantId := azure.GetSubscriptionAndTenantId(account)
			Expect(id).To(Equal("some-id"))
			Expect(tenantId).To(Equal("some-tenant-id"))
		})
	})

	Describe("AppExists", func() {
		Context("when no applications with that display name exist", func() {
			BeforeEach(func() {
				cli.ExecuteCall.Returns.Output = `[]`
			})

			It("returns no error", func() {
				err := azure.AppExists()
				Expect(err).NotTo(HaveOccurred())

				Expect(cli.ExecuteCall.Receives.Args).To(Equal([]string{"ad", "app", "list", "--display-name", "some-display-name"}))
				Expect(logger.PrintlnCall.Receives.Message).To(Equal("Confirmed no application already exists with display name."))
			})
		})

		Context("when an application with that display name exists", func() {
			BeforeEach(func() {
				cli.ExecuteCall.Returns.Output = `[{"displayName": "some-display-name", "appId": "1234"}]`
			})

			It("returns a helpful error", func() {
				err := azure.AppExists()
				Expect(err).To(MatchError("The --display-name some-display-name is taken by application with id 1234."))
			})
		})

		Context("when the cli returns an error", func() {
			BeforeEach(func() {
				cli.ExecuteCall.Returns.Error = errors.New("some error")
				cli.ExecuteCall.Returns.Output = "the error message"
			})

			It("returns a helpful error", func() {
				err := azure.AppExists()
				Expect(err).To(MatchError("Running [ad app list --display-name some-display-name]: the error message"))
			})
		})

		Context("when the applications json is invalid", func() {
			BeforeEach(func() {
				cli.ExecuteCall.Returns.Output = `[{$$$}]`
			})

			It("returns a helpful error", func() {
				err := azure.AppExists()
				Expect(err).To(MatchError(ContainSubstring("Unmarshalling applications json: ")))
			})
		})
	})

	Describe("CreateApplication", func() {
		var clientSecret string
		BeforeEach(func() {
			clientSecret = "the-client-secret"
			cli.ExecuteCall.Returns.Output = `{"appId": "the-client-id"}`
		})

		It("returns the client id and client secret", func() {
			clientId, err := azure.CreateApplication(clientSecret)
			Expect(err).NotTo(HaveOccurred())

			Expect(cli.ExecuteCall.Receives.Args).To(Equal([]string{"ad", "app", "create",
				"--display-name", "some-display-name",
				"--homepage", "http://some-identifier-uri",
				"--identifier-uris", "http://some-identifier-uri",
				"--password", "the-client-secret",
			}))
			Expect(clientId).To(Equal("the-client-id"))
			Expect(logger.PrintlnCall.Receives.Message).To(Equal("Created application."))
		})

		Context("when the cli returns an error", func() {
			BeforeEach(func() {
				cli.ExecuteCall.Returns.Error = errors.New("some error")
			})

			It("returns a helpful error", func() {
				_, err := azure.CreateApplication(clientSecret)
				Expect(err).To(MatchError(ContainSubstring("Running [ad app create --display-name some-display-name")))
				Expect(err).NotTo(MatchError(ContainSubstring("--password the-client-secret")))
			})
		})

		Context("when the application json is invalid", func() {
			BeforeEach(func() {
				cli.ExecuteCall.Returns.Output = `{$$$}`
			})

			It("returns a helpful error", func() {
				_, err := azure.CreateApplication(clientSecret)
				Expect(err).To(MatchError(ContainSubstring("Unmarshalling application json: ")))
			})
		})
	})

	Describe("CreateServicePrincipal", func() {
		It("creates the service principal", func() {
			err := azure.CreateServicePrincipal("the-client-id")
			Expect(err).NotTo(HaveOccurred())

			Expect(cli.ExecuteCall.Receives.Args).To(Equal([]string{"ad", "sp", "create", "--id", "the-client-id"}))
			Expect(logger.PrintlnCall.Receives.Message).To(Equal("Created service principal."))
		})

		Context("when the cli returns an error", func() {
			BeforeEach(func() {
				cli.ExecuteCall.Returns.Error = errors.New("some error")
			})

			It("returns a helpeful error", func() {
				err := azure.CreateServicePrincipal("the-client-id")
				Expect(err).To(MatchError(ContainSubstring("Running [ad sp create --id the-client-id]: ")))
			})
		})
	})

	Describe("AssignContributorRole", func() {
		It("assigns the contributor role to the service principal", func() {
			err := azure.AssignContributorRole("the-client-id")
			Expect(err).NotTo(HaveOccurred())

			Expect(cli.ExecuteCall.Receives.Args).To(Equal([]string{"role", "assignment", "create",
				"--role", "Contributor",
				"--assignee", "the-client-id"}))
			Expect(logger.PrintlnCall.Receives.Message).To(Equal("Assigned contributor role to service principal."))
		})

		Context("when the cli returns an error", func() {
			BeforeEach(func() {
				cli.ExecuteCall.Returns.Error = errors.New("some error")
			})

			It("returns a helpeful error", func() {
				err := azure.AssignContributorRole("the-client-id")
				Expect(err).To(MatchError(ContainSubstring("Running [role assignment create --role Contributor --assignee the-client-id]: ")))
			})
		})
	})

	Describe("WriteCredentials", func() {
		AfterEach(func() {
			err := os.Remove("some-credential-file")
			Expect(err).NotTo(HaveOccurred())
		})

		It("writes the credentials to the specified output file", func() {
			err := azure.WriteCredentials("subscription-id", "tenant-id", "client-id", "client-secret")
			Expect(err).NotTo(HaveOccurred())

			bytes, err := ioutil.ReadFile(credentialOutputFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(bytes)).To(ContainSubstring("subscription_id = subscription-id"))
			Expect(string(bytes)).To(ContainSubstring("tenant_id = tenant-id"))
			Expect(string(bytes)).To(ContainSubstring("client_id = client-id"))
			Expect(string(bytes)).To(ContainSubstring("client_secret = client-secret"))

			Expect(logger.PrintlnCall.Receives.Message).To(Equal("Wrote credentials to some-credential-file."))
		})
	})
})
