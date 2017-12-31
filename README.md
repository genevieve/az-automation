# Create an azure automation account

Requirements:

- [Azure CLI](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli?view=azure-cli-latest)

Steps:

1. Log in to the azure cli

    ```
    az login
    az account list
    ```

1. Run

    ```
    az-automation
      --acount your-account-name \
      --identifier-uri http://example.com \
      --display-name example-applicaion-name \
      --credential-output-file creds.tfvars
    ```
