package main

import (
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/genevieve/az-automation/az"
	flags "github.com/jessevdk/go-flags"
)

type args struct {
	Account              string `required:"true" short:"a" long:"account"                description:"Your account id or name. Use 'az account list' to see your accounts."`
	DisplayName          string `required:"true" short:"d" long:"display-name"           description:"Display name for application. Must be unique."`
	IdentifierUri        string `required:"true" short:"i" long:"identifier-uri"         description:"Must be unique."`
	CredentialOutputFile string `required:"true" short:"c" long:"credential-output-file" description:"Must be unique."                                                      default:"creds.tfvars"`
}

func main() {
	log.SetFlags(0)

	var a args
	parser := flags.NewParser(&a, flags.HelpFlag|flags.PrintErrors)
	_, err := parser.ParseArgs(os.Args)
	if err != nil {
		os.Exit(0)
	}

	path, err := exec.LookPath("az")
	if err != nil {
		log.Fatalf("Failed to find the azure-cli (`az`): %s", err)
	}

	cli := az.NewCLI(path)
	logger := az.NewLogger(os.Stdout)
	azure := az.NewAz(cli, logger)

	err = azure.ValidVersion()
	if err != nil {
		log.Fatal(err)
	}

	account, err := azure.LoggedIn(a.Account)
	if err != nil {
		log.Fatal(err)
	}

	err = azure.AppExists(a.DisplayName)
	if err != nil {
		log.Fatal(err)
	}

	clientSecret := azure.GeneratePassword()
	clientId, err := azure.CreateApplication(clientSecret, a.DisplayName, a.IdentifierUri)
	if err != nil {
		log.Fatal(err)
	}

	err = azure.CreateServicePrincipal(clientId)
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(30 * time.Second)

	err = azure.AssignContributorRole(clientId)
	if err != nil {
		log.Fatal(err)
	}

	id, tenantId := azure.GetSubscriptionAndTenantId(account)
	err = azure.WriteCredentials(id, tenantId, clientId, clientSecret, a.CredentialOutputFile)
	if err != nil {
		log.Fatal(err)
	}
}
