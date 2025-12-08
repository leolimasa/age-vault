* Implement the new `vault-key from-identity` command to initialize the vault key from the identity file as specified in the `README.md`.
* Fix the age_vault.yml configuration mechanism so that the relative paths are relative to the config file location, not the current working directory.
* When the ssh agent is running, add logging to indicate which keys are being loaded from the vault, when it uses a key to sign a request, and any errors encountered during these operations.
* Ensure the ssh agent always reloads keys from the vault directory when a signing request is received, to pick up any new keys added since the last request.
* On test/, create a `start_ssh_server.sh` that runs a localhost only temporary ssh daemon that will accept the public key specified as the first argument for authentication.
* On test/, create a new run_integration_tests.sh script that sets up a temporary vault directory using `age` and the compiled `age-vault` command. Then, run all possible workflows using a compiled `age-vault`, including starting up the ssh agent and testing the connection to the docker server created above.
