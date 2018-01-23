# Create an azure automation account

Requirements:

- [Azure CLI](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli?view=azure-cli-latest)

## Installation

```
brew tap genevievelesperance/tap
brew install az-automation
```

## Usage


```
Usage:
  az-automation [OPTIONS]

Application Options:
  -a, --account=                Your account id or name. Use 'az account list' to see your accounts.
  -d, --display-name=           Display name for application. Must be unique.
  -i, --identifier-uri=         Must be unique.
  -c, --credential-output-file= Must be unique. (default: creds.tfvars)

Help Options:
  -h, --help                    Show this help message
```


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
