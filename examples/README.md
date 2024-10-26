# Pipeline

The pipeline utilizes two tokens, both stored as CI/CD variables with the [hidden](https://docs.gitlab.com/ee/ci/variables/#hide-a-cicd-variable) option for enhanced security:

- **`TOKEN_RO`**: A read-only token with the `read_api` scope, used exclusively in the test stage (e.g., during the merge request phase). This token allows testing with limited permissions, preventing unintended changes.
- **`TOKEN_RW`**: A read-write token with the `api` scope, used for all primary executions, including scheduled tasks and when the force option is enabled. This variable is restricted to [protected branches](https://docs.gitlab.com/ee/ci/variables/#protect-a-cicd-variable), ensuring it is accessible only within secure branch contexts.
