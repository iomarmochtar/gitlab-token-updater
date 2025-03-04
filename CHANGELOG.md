# 0.4.0

### Features and enhancements

- [hook] update variable contents in another Gitlab instance
- [hook] evaluting env var character in hook args for `update_var` and `.exec_cmd` (`.envs`)

### Bug Fixes

- non exists access token in configuration not treated as an error
- processing inactive token

# 0.3.0

### Features and enhancements

- [config] set hook arguments type and path config for `update_var`


# 0.2.0

### Features and enhancements

- [config] read external config file for `manage_tokens` by keyword `include`
- [config] add validation in detecting duplicated manage token
- [config] give the relevan sequence of errors detected in `manage_tokens` config, for easier in config trouble shooting
- [log] introduce arg `--log-json` for logging in JSON format
- [hook] passing additional env var in `exec_cmd` by args `env`

# 0.1.1

### Bug Fixes

- handling once the returned access token is without expires

# 0.1.0

### Features and enhancements

- configuration validations, tested with multiple scenarios in unit test
- load configuration for `host` and `token` from environment variable
- global/default configuration for `renew_before` and `expiry_before_rotate` in each access token config, `hook_retry` on hook config
- self rotate the used token for access token by type `personal` and switch it with the new one afterward by using hook `use_token`
- update group or project CICD variable after the rotation by using hook `update_var`
- run the executeable (script/binary file) to accomodate other use cases in by using hook `exec_cmd`
- run dry run mode using argument `--dry-run` in validating configuration with the actual one
- strict mode using argument `--strict` for not continuing the execution once an error occured
- running in force mode using argument `--force` in executing token rotation regardless it's expires time 
