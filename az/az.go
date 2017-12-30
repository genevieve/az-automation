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
	cli CLI

	creds credentials

	account              string
	displayName          string
	identifierUri        string
	credentialOutputFile string
}

type credentials struct {
	SubscriptionId string
	TenantId       string
	ClientId       string
	ClientSecret   string
}

func NewAz(cli CLI, account, displayName, identifierUri, credentialOutputFile string) *Az {
	return &Az{
		cli:                  cli,
		account:              account,
		displayName:          displayName,
		identifierUri:        identifierUri,
		credentialOutputFile: credentialOutputFile,
	}
}

func (a *Az) ValidVersion() error {
	output, err := a.cli.Execute([]string{"-v"})
	if err != nil {
		return errors.New("Please install the azure-cli.")
	}

	regex := regexp.MustCompile(`\d+.\d+.\d+`)
	v := regex.FindString(output)

	curr, err := semver.NewVersion(v)
	if err != nil {
		return errors.New("The azure-cli version (`az -v`) could not be parsed.")
	}

	min, _ := semver.NewVersion("2.0.0")

	if curr.LessThan(min) {
		return errors.New("Please update your azure-cli to at least 2.0.0.")
	}

	return nil
}

func (a *Az) LoggedIn() error {
	output, err := a.cli.Execute([]string{"account", "list"})
	if err != nil {
		return errors.New("Please login in to the azure-cli.")
	}

	accounts := []Account{}
	err = json.Unmarshal([]byte(output), &accounts)
	if err != nil {
		return errors.New(fmt.Sprintf("Unmarshalling accounts json: %s", err))
	}

	if len(accounts) == 0 {
		return errors.New("Login to the azure-cli (`az login`).")
	}

	return nil
}

func (a *Az) GetSubscriptionAndTenantId() error {
	output, err := a.cli.Execute([]string{"account", "show", "-s", a.account})
	if err != nil {
		return err
	}

	account := Account{}
	err = json.Unmarshal([]byte(output), &account)
	if err != nil {
		return errors.New(fmt.Sprintf("Unmarshalling account json: %s", err))
	}

	a.creds.SubscriptionId = account.Id
	a.creds.TenantId = account.TenantId

	return nil
}

func (a *Az) AppExists() error {
	output, err := a.cli.Execute([]string{
		"ad", "app", "list",
		"--display-name", a.displayName,
	})
	if err != nil {
		return errors.New(fmt.Sprintf("Running `az ad app list`: %s", err))
	}

	applications := []Application{}
	err = json.Unmarshal([]byte(output), &applications)
	if err != nil {
		return errors.New(fmt.Sprintf("Unmarshalling applications json: %s", err))
	}

	if len(applications) > 0 {
		return errors.New(fmt.Sprintf("The --display name %s is taken by application with id %s.", a.displayName, applications[0].AppId))
	}

	return nil
}

func (a *Az) CreateApplication() error {
	a.creds.ClientSecret = uuid.Must(uuid.NewRandom()).String()

	createArgs := []string{
		"ad", "app", "create",
		"--display-name", a.displayName,
		"--homepage", a.identifierUri,
		"--identifier-uris", a.identifierUri,
	}

	output, err := a.cli.Execute(append(createArgs, "--password", a.creds.ClientSecret))
	if err != nil {
		return errors.New(fmt.Sprintf("Running %+v: %s", createArgs, output))
	}

	application := Application{}
	err = json.Unmarshal([]byte(output), &application)
	if err != nil {
		return errors.New(fmt.Sprintf("Unmarshalling application json: %s", err))
	}

	a.creds.ClientId = application.AppId

	return nil
}

func (a *Az) CreateServicePrincipal() error {
	args := []string{
		"ad", "sp", "create",
		"--id", a.creds.ClientId,
	}

	output, err := a.cli.Execute(args)
	if err != nil {
		return errors.New(fmt.Sprintf("Running %+v: %s", args, output))
	}

	return nil
}

func (a *Az) AssignContributorRole() error {
	args := []string{
		"role", "assignment", "create",
		"--role", "Contributor",
		"--assignee", a.creds.ClientId,
		"--scope", fmt.Sprintf("/subscriptions/%s", a.creds.SubscriptionId),
	}

	output, err := a.cli.Execute(args)
	if err != nil {
		return errors.New(fmt.Sprintf("Running %+v: %s", args, output))
	}

	return nil
}

func (a *Az) WriteCredentials() error {
	creds := fmt.Sprintf(`subscription_id = %s
tenant_id = %s
client_id = %s
client_secret = %s
`,
		a.creds.SubscriptionId,
		a.creds.TenantId,
		a.creds.ClientId,
		a.creds.ClientSecret)

	err := ioutil.WriteFile(a.credentialOutputFile, []byte(creds), 0600)
	if err != nil {
		return errors.New(fmt.Sprintf("Writing credentials to output file: %s", err))
	}

	return nil
}
