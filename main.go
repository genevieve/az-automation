package main

import (
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/genevievelesperance/az-automation/az"
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
	azure := az.NewAz(cli, a.Account, a.DisplayName, a.IdentifierUri, a.CredentialOutputFile)

	err = azure.ValidVersion()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Checked version of azure-cli.")

	account, err := azure.LoggedIn()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Checked you are logged in to azure-cli (`az`).")

	id, tenantId := azure.GetSubscriptionAndTenantId(account)
	log.Println("Retrieved subscription and tenant id.")

	err = azure.AppExists()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Confirmed application name is not already taken.")

	clientSecret := azure.GeneratePassword()
	clientId, err := azure.CreateApplication(clientSecret)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Created application.")

	log.Println("Creating service principal.")
	err = azure.CreateServicePrincipal(clientId)
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(30 * time.Second)
	log.Println("Created service principal.")

	err = azure.AssignContributorRole(clientId)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Assigned contributor role to service principal.")

	err = azure.WriteCredentials(id, tenantId, clientId, clientSecret)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Wrote credentials to output file.")
}
