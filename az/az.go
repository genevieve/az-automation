package az

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"

	"github.com/google/uuid"
	semver "github.com/hashicorp/go-version"
)

type Account struct {
	Name     string `json:"name"`
	Id       string `json:"id"`
	TenantId string `json:"tenantId"`
}

type Application struct {
	DisplayName string `json:"displayName"`
	AppId       string `json:"appId"`
}

type ServicePrincipal struct {
	AppId string `json:"appId"`
}

type Az struct {
	cli                  cli
	logger               logger
	account              string
	displayName          string
	identifierUri        string
	credentialOutputFile string
}

type cli interface {
	Execute(args []string) (string, error)
}

type logger interface {
	Println(message string)
}

func NewAz(cli cli, logger logger, account, displayName, identifierUri, credentialOutputFile string) *Az {
	return &Az{
		cli:                  cli,
		logger:               logger,
		account:              account,
		displayName:          displayName,
		identifierUri:        identifierUri,
		credentialOutputFile: credentialOutputFile,
	}
}

func (a Az) ValidVersion() error {
	output, err := a.cli.Execute([]string{"-v"})
	if err != nil {
		return errors.New("Please install the azure-cli.")
	}

	regex := regexp.MustCompile(`\d+.\d+.\d+`)
	v := regex.FindString(output)

	curr, err := semver.NewVersion(v)
	if err != nil {
		return errors.New("The azure-cli version could not be parsed.")
	}

	min, _ := semver.NewVersion("2.0.0")

	if curr.LessThan(min) {
		return errors.New("Please update the azure-cli to at least 2.0.0.")
	}

	a.logger.Println("Checked version of azure-cli is above 2.0.0.")
	return nil
}

func (a Az) LoggedIn() (Account, error) {
	account := Account{}

	output, err := a.cli.Execute([]string{"account", "show", "-s", a.account})
	if err != nil {
		return account, errors.New("Please login to the azure-cli.")
	}

	err = json.Unmarshal([]byte(output), &account)
	if err != nil {
		return account, errors.New(fmt.Sprintf("Unmarshalling account json: %s", err))
	}

	a.logger.Println("Checked you are logged in to the azure-cli.")
	return account, nil
}

func (a Az) GetSubscriptionAndTenantId(account Account) (string, string) {
	return account.Id, account.TenantId
}

func (a Az) AppExists() error {
	args := []string{
		"ad", "app", "list",
		"--display-name", a.displayName,
	}

	output, err := a.cli.Execute(args)
	if err != nil {
		return errors.New(fmt.Sprintf("Running %+v: %s", args, output))
	}

	applications := []Application{}
	err = json.Unmarshal([]byte(output), &applications)
	if err != nil {
		return errors.New(fmt.Sprintf("Unmarshalling applications json: %s", err))
	}

	if len(applications) > 0 {
		return errors.New(fmt.Sprintf("The --display-name %s is taken by application with id %s.", a.displayName, applications[0].AppId))
	}

	a.logger.Println("Confirmed no application already exists with display name.")
	return nil
}

func (a Az) GeneratePassword() string {
	return uuid.Must(uuid.NewRandom()).String()
}

func (a Az) CreateApplication(password string) (string, error) {
	createArgs := []string{
		"ad", "app", "create",
		"--display-name", a.displayName,
		"--homepage", a.identifierUri,
		"--identifier-uris", a.identifierUri,
	}

	output, err := a.cli.Execute(append(createArgs, "--password", password))
	if err != nil {
		return "", errors.New(fmt.Sprintf("Running %+v: %s", createArgs, output))
	}

	application := Application{}
	err = json.Unmarshal([]byte(output), &application)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Unmarshalling application json: %s", err))
	}

	a.logger.Println("Created application.")
	return application.AppId, nil
}

func (a Az) CreateServicePrincipal(clientId string) error {
	createArgs := []string{
		"ad", "sp", "create",
		"--id", clientId,
	}

	output, err := a.cli.Execute(createArgs)
	if err != nil {
		return errors.New(fmt.Sprintf("Running %+v: %s", createArgs, output))
	}

	a.logger.Println("Created service principal.")
	return nil
}

func (a Az) AssignContributorRole(clientId string) error {
	args := []string{
		"role", "assignment", "create",
		"--role", "Contributor",
		"--assignee", clientId,
	}

	output, err := a.cli.Execute(args)
	if err != nil {
		return errors.New(fmt.Sprintf("Running %+v: %s", args, output))
	}

	a.logger.Println("Assigned contributor role to service principal.")
	return nil
}

func (a *Az) WriteCredentials(id, tenantId, clientId, clientSecret string) error {
	creds := fmt.Sprintf(`subscription_id = %s
tenant_id = %s
client_id = %s
client_secret = %s
`,
		id,
		tenantId,
		clientId,
		clientSecret)

	err := ioutil.WriteFile(a.credentialOutputFile, []byte(creds), 0600)
	if err != nil {
		return errors.New(fmt.Sprintf("Writing credentials to output file: %s", err))
	}

	a.logger.Println(fmt.Sprintf("Wrote credentials to %s.", a.credentialOutputFile))
	return nil
}
