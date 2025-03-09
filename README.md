# GitLab Access Token Updater/Rotation Manager

This app addresses the following challenges:

- **Token Expiration Management**: If you've created numerous Group/Project Access Tokens without expiration dates, these tokens can unexpectedly become a problem once got expired (broken automation, pipeline, etc), as outlined in [GitLab's new expiration policy](https://about.gitlab.com/blog/2023/10/25/access-token-lifetime-limits/).
- **Premium Solution Limitations**: GitLab Premium or Ultimate users are recommended to use Service Account personal access tokens that don’t expire (mentioned in blog above). However, this introduces additional security risks.
- **Service Account Management**: Not a direct issue, managing Service Accounts is difficult due to the lack of a UI, as noted in [GitLab's current limitations](https://gitlab.com/groups/gitlab-org/-/epics/9965), requiring manual API calls for management.
- **Token Usage Visibility**: Tracking Group/Project access token usage is limited, and rotating or renewing tokens requires locating and updating each token’s use, such as in CI/CD or Kubernetes secrets.
- **Response to Token Leaks or Staff Changes**: When tokens are leaked or an employee leaves, the lack of visibility or automation around access tokens makes you don't want rotate it.

## Installation

### Docker

Assuming your main configuration file is located in the current directory:

```bash
docker run -v `./config.yml:/tmp/config.yml` -it --rm iomarmochtar/gitlab-token-updater:latest -c /tmp/config.yml
```

### Compiled

Go to the releases page for a compiled version specific to your OS and architecture.

### Source Build

Ensure to have following package installed on your local:
- Golang binary with the same version as in [go.mod](./go.mod)
- [goreleaser](https://goreleaser.com/install/), alternatively, use script [install_goreleaser.sh](./scripts/install_goreleaser.sh) to automate it
- [GNU Make](https://www.gnu.org/software/make/)

Run following commands:

```bash
git clone https://github.com/iomarmochtar/gitlab-token-updater.git
cd gitlab-token-updater
make dist-dev
```

the compiled binary is located under folder `dist`


## How To Use

### Preparations

#### 1. Personal Access Token

As Group/Project Access Tokens cannot be used to [create other access token](https://docs.gitlab.com/ee/user/group/settings/group_access_tokens.html) related to the access token rotation API ([example](https://docs.gitlab.com/ee/api/project_access_tokens.html#rotate-a-project-access-token)), use a Personal Access Token for access token rotation.

Create a Personal Access Token with the api scope. You may also create another with the read_api scope for MR dry run purposes.

#### 2. Config File

To run the application, specify the configuration file with the `-c` or `--config` argument. Other modes are available.

### Execution

For the normal execution, just poin the configuration file with argument `-c` or `--config`, some of other modes are available

```bash
gitlab-token-updater -c [PATH_TO_CONFIG_FILE]
```

#### Mode

You can enable each mode by adding the respective argument, and they can be used in combination.

If any error occurs during execution, it will not interrupt the iteration; errors are accumulated and reported at the end. This behavior can be modified by running in [strict mode](#strict).

##### Dry Run

Run with the `--dry-run` argument. This will execute read-only APIs and verify the existence of specified objects like access tokens and environment variables.

It can be combined with [force mode](#force) to scan for the existence of each hook. This is recommended during MR validation.

##### Force

Run with the `--force` argument to execute the rotation API regardless of the access token's expiry time. This acts as an "emergency button" to ensure all tokens are renewed.

##### Strict

Run with the `--strict` argument. Any error encountered during execution will be raised immediately, stopping the process.

## Configuration

Consist of YAML formatted content, see the sample one in [sample-config.yml](./examples/sample-config.yml), these are the available properties

| Param                                                  | Description                                                                                                 | Defaults              |             Required              |
| ------------------------------------------------------ | ----------------------------------------------------------------------------------------------------------- | --------------------- | :-------------------------------: |
| `.host`                                                | URL of Gitlab instance                                                                                      | `https://gitlab.com/` |               `yes`               |
| `.token`                                               | Personal access token for GitLab API usage (api scope for all execution, read_api for dry run mode)         | `${GL_RENEWER_TOKEN}` |               `yes`               |
| `.default_hook_retry`                                  | Default retry count for hook execution; can be overridden in individual hook configurations                 | `0`                   |               `yes`               |
| `.default_renew_before`                                | Default duration to renew an access token before expiry; can be overridden in specific access token configs | `14d`                 |               `yes`               |
| `.default_expiry_after_rotate`                         | Default duration for token expiration after rotation                                                        | `3M`                  |               `yes`               |
| `.manage_tokens[]`                                     | List of managed access token                                                                                |                       |               `yes`               |
| `.manage_tokens[].type`                                | Type of access token (`repository`, `group`, or `personal`)                                                 |                       |               `yes`               |
| `.manage_tokens[].path`                                | Repository or group location                                                                                |                       | Required for `repository`/`group` |
| `.manage_tokens[].include`                             | Include external `manage_token` configuration, the path is relative to main config file                     |                       |               `no`                |
| `.manage_tokens[].access_tokens[]`                     | List of managed access tokens                                                                               |                       |               `yes`               |
| `.manage_tokens[].access_tokens[].name`                | Name of access token                                                                                        |                       |               `yes`               |
| `.manage_tokens[].access_tokens[].renew_before`        | Specific renewal period, overriding `.default_renew_before`                                                 |                       |               `no`                |
| `.manage_tokens[].access_tokens[].expiry_after_rotate` | Specific expiration period, overriding `.default_expiry_after_rotate`                                       |                       |               `no`                |
| `.manage_tokens[].access_tokens[].hooks[]`             | List of actions for each hook                                                                               |                       |               `no`                |
| `.manage_tokens[].access_tokens[].hooks[].type`        | Hook type (`update_var`, `exec_cmd`, `use_token`)                                                           |                       |               `yes`               |
| `.manage_tokens[].access_tokens[].hooks[].retry`       | Hook retry count, overriding `.default_hook_retry`                                                          |                       |               `no`                |
| `.manage_tokens[].access_tokens[].hooks[].args`        | Arguments for each hook type (see details below)                                                            |                       |  *some hook type is not required  |

**Notes:**

- the config value can be injected by env variable by using format `${THIS_IS_VAR}` and it's only available in following locations:
  - `.host`
  - `.token`
  - `.manage_tokens[].access_tokens[].hooks[].args` for hook type `update_var`
  - `.manage_tokens[].access_tokens[].hooks[].args.env` for hook type `exec_cmd`
- Known duration suffixes: `d` (day), `M` (month), `Y` (year).
- hook types with it's available arguments:
  - `update_var`:
    - `.name` (required): CICD variable name
    - `.type` (required): `repository` or `group`
    - `.path` (required): location of repository or group
    - `.gitlab`: set this if CICD variable is located in another Gitlab instance
    - `.gitlab_token`: the token that will be used to another Gitlab instance as in `.gitlab`
    - misc:
      - if `.type` and `.path` are not defined, then it will use the same as in it's parent (manage token config)
      - `.gitlab-token` is required when `.gitlab` configured, and suggested set in env variable
  - `exec_cmd`:
    - `.path` (required): location of executable
    - `.env`: set the injected environment variable that will be read by the executeable
    - misc:
      - the new generated token that read in the executeable is through env variable by name `GL_NEW_TOKEN`
  - `use_token`: not requiring any arguments, it will uses the new token in the current API call; can only be set once in the first hook sequence.

## Development

To avoid "polluting" your local environment and to use a consistent development setup, use [devcontainer](https://containers.dev/), which is included in this repository and a built in feature in Visual Studio Code.

## Misc

### Utilizing Service Account

For GitLab Premium or higher, it's recommended to use a Service Account's access token instead of one belonging to an active user. Consider leaving its expiration unset and using a self-update approach, as shown in the [examples folder](./examples/).

### Separating Between Access Token For Execution

For more information, refer to [this example](./examples/README.md).