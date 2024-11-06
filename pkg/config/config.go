// Package config all related to configuration, it also include with content validations
package config

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	defaultHost              = "https://gitlab.com/"
	defaultToken             = "${GL_RENEWER_TOKEN}"
	defaultRenewBefore       = "14d"
	defaultExpiryAfterRotate = "3M"
	defaultHookRetry         = 0
	ManagedTypeRepository    = "repository"
	ManagedTypeGroup         = "group"
	ManagedTypePersonal      = "personal"
	HookTypeUpdateVar        = "update_var"
	HookTypeExecCMD          = "exec_cmd"
	HookTypeUseToken         = "use_token"
)

var (
	ManagedTypeList = []string{
		ManagedTypePersonal,
		ManagedTypeGroup,
		ManagedTypeRepository,
	}
	HookTypeList = []string{
		HookTypeUpdateVar,
		HookTypeExecCMD,
		HookTypeUseToken,
	}
)

var (
	ErrValidationEmptyManagedList                = errors.New("empty managed list")
	ErrValidationEmptyGitlabToken                = errors.New("empty gitlab token")
	ErrValidationEmptyHost                       = errors.New("empty host")
	ErrValidationInvalidDefaultRenewBefore       = errors.New("invalid default renew before value")
	ErrValidationInvalidDefaultExpiryAfterRotate = errors.New("invalid default expiry after rotate value")
	ErrValidationManagedEmptyPath                = errors.New("empty path config in managed token")
	ErrValidationManagedInvalidType              = fmt.Errorf("invalid type, the valid one are %s", strings.Join(ManagedTypeList, ","))
	ErrValidationManagedEmptyTokenList           = errors.New("empty managed token list")
	ErrValidationManagedInvalidRenewBefore       = errors.New("invalid renew before value")
	ErrValidationManagedInvalidExpiryAfterRotate = errors.New("invalid expiry after rotate value")
	ErrValidationManagedDuplicatedDefinition     = errors.New("duplicated manage token found")
	ErrValidationTokenEmptyName                  = errors.New("empty token name")
	ErrValidationHookInvalidType                 = fmt.Errorf("invalid hook type, the valid one are %s", strings.Join(HookTypeList, ","))
	ErrValidationHookUpdateVarMissingName        = fmt.Errorf("missing arg name in %s hook", HookTypeUpdateVar)
	ErrValidationHookUpdateVarMissingPath        = fmt.Errorf("missing arg path in %s hook", HookTypeUpdateVar)
	ErrValidationHookUpdateVarMissingType        = fmt.Errorf("missing arg type in %s hook", HookTypeUpdateVar)
	ErrValidationHookUpdateVarInvalidType        = fmt.Errorf("invalid arg type in %s hook, the valid one are %s", HookTypeUpdateVar, strings.Join(ManagedTypeList, ","))
	ErrValidationHookExecCMDMissingPath          = fmt.Errorf("missing arg path in %s hook", HookTypeExecCMD)
	ErrValidationHookUseTokenNotByPersonalType   = fmt.Errorf("can be only use in manage type %s", ManagedTypePersonal)
	ErrValidationHookUseTokenAlreadyUse          = fmt.Errorf("hook %s can be only use once", HookTypeUseToken)
	ErrValidationHookUseTokenNotFirstSeq         = fmt.Errorf("hook %s must be set at the first", HookTypeUseToken)
)

type HookUpdateVar struct {
	Name string
	Path string
	Type string
}

type HookExecScript struct {
	Path string
}

type Hook struct {
	Type  string            `yaml:"type"`
	Retry uint8             `yaml:"retry"`
	Args  map[string]string `yaml:"args"`
}

func (h Hook) validate() error {
	if !contains(HookTypeList, h.Type) {
		return ErrValidationHookInvalidType
	}

	if h.Type == HookTypeUpdateVar {
		if h.Args["name"] == "" {
			return ErrValidationHookUpdateVarMissingName
		}

		if h.Args["path"] == "" {
			return ErrValidationHookUpdateVarMissingPath
		}

		if !contains(ManagedTypeList, h.Args["type"]) {
			return ErrValidationHookUpdateVarInvalidType
		}
	} else if h.Type == HookTypeExecCMD {
		if h.Args["path"] == "" {
			return ErrValidationHookExecCMDMissingPath
		}
	}
	return nil
}

func (h Hook) UpdateVarArgs() HookUpdateVar {
	return HookUpdateVar{
		Name: h.Args["name"],
		Path: h.Args["path"],
		Type: h.Args["type"],
	}
}

func (h Hook) ExecCMDArgs() HookExecScript {
	return HookExecScript{
		Path: h.Args["path"],
	}
}

func (h Hook) StrArgs() string {
	switch h.Type {
	case HookTypeUpdateVar:
		args := h.UpdateVarArgs()
		return fmt.Sprintf("type:%s,path:%s,name:%s", args.Type, args.Path, args.Name)
	case HookTypeExecCMD:
		args := h.ExecCMDArgs()
		return fmt.Sprintf("path:%s", args.Path)
	}
	return ""
}

type AccessToken struct {
	Name              string `yaml:"name"`
	RenewBefore       string `yaml:"renew_before"`
	ExpiryAfterRotate string `yaml:"expiry_after_rotate"`
	Hooks             []Hook `yaml:"hooks"`
}

func (at AccessToken) RenewBeforeDuration() (time.Duration, error) {
	return durationParse(at.RenewBefore)
}

func (at AccessToken) ExpiryAfterRotateDuration() (time.Duration, error) {
	return durationParse(at.ExpiryAfterRotate)
}

func (at AccessToken) validate() error {
	if at.Name == "" {
		return ErrValidationTokenEmptyName
	}

	if at.RenewBefore != "" && !renewBeforeRe.MatchString(at.RenewBefore) {
		return ErrValidationManagedInvalidRenewBefore
	}

	if at.ExpiryAfterRotate != "" && !renewBeforeRe.MatchString(at.ExpiryAfterRotate) {
		return ErrValidationManagedInvalidExpiryAfterRotate
	}
	return nil
}

type ManagedToken struct {
	Path   string        `yaml:"path"`
	Type   string        `yaml:"type"`
	Ref    string        `yaml:"include"`
	Tokens []AccessToken `yaml:"access_tokens"`
}

func (m *ManagedToken) validate() error {
	if !contains(ManagedTypeList, m.Type) {
		return ErrValidationManagedInvalidType
	}

	// except personal_token path property not required
	if m.Path == "" && m.Type != ManagedTypePersonal {
		return ErrValidationManagedEmptyPath
	}

	if len(m.Tokens) == 0 {
		return ErrValidationManagedEmptyTokenList
	}

	return nil
}

type Config struct {
	Host                     string         `yaml:"host"`
	Token                    string         `yaml:"token"`
	DefaultHookRetry         uint8          `yaml:"default_hook_retry"`
	DefaultRenewBefore       string         `yaml:"default_renew_before"`
	DefaultExpiryAfterRotate string         `yaml:"default_expiry_after_rotate"`
	Managed                  []ManagedToken `yaml:"manage_tokens"`
}

func (c Config) DefaultRenewBeforeDuration() (time.Duration, error) {
	return durationParse(c.DefaultRenewBefore)
}

func (c Config) DefaultExpiryAfterRotateDuration() (time.Duration, error) {
	return durationParse(c.DefaultExpiryAfterRotate)
}

// appendReference adding more context to returned error
func appendErrReferences(err error, references []string) error {
	for _, errRef := range references {
		err = errors.Join(err, errors.New(errRef))
	}
	return err
}

func (c Config) validate() (err error) {
	if c.Host == "" {
		return ErrValidationEmptyHost
	}

	if len(c.Managed) == 0 {
		return ErrValidationEmptyManagedList
	}

	if c.Token == "" {
		return ErrValidationEmptyGitlabToken
	}

	if _, err = c.DefaultRenewBeforeDuration(); err != nil {
		return errors.Join(ErrValidationInvalidDefaultRenewBefore, err)
	}

	if _, err = c.DefaultExpiryAfterRotateDuration(); err != nil {
		return errors.Join(ErrValidationInvalidDefaultExpiryAfterRotate, err)
	}

	hookUseTokenUsed := false
	// track sequence number of managed_token
	managedRefSeq := make(map[string]int)
	// track the used manage token
	managedTrackUsed := make(map[string]string)

	for idx := range c.Managed {
		managed := c.Managed[idx]
		var errRefsManage []string
		if managed.Ref != "" {
			errRefsManage = append(errRefsManage, fmt.Sprintf("reference: %s", managed.Ref))
		}

		if _, exists := managedRefSeq[managed.Ref]; !exists {
			managedRefSeq[managed.Ref] = 0
		}
		managedRefSeq[managed.Ref]++
		info := fmt.Sprintf("managed_token seq num: %d", managedRefSeq[managed.Ref])
		if managed.Type == ManagedTypePersonal {
			info = fmt.Sprintf("%s (type: %s)", info, ManagedTypePersonal)
		} else {
			info = fmt.Sprintf("%s (type: %s, path: %s)", info, managed.Type, managed.Path)
		}
		errRefsManage = append(errRefsManage, info)

		managedID := fmt.Sprintf("mtkn_%s_%s", managed.Type, managed.Path)
		prevManageRef, exists := managedTrackUsed[managedID]
		if !exists {
			managedTrackUsed[managedID] = managed.Ref
		} else {
			err = ErrValidationManagedDuplicatedDefinition
			if prevManageRef != "" {
				err = errors.Join(err, fmt.Errorf("previously defined at %s", prevManageRef))
			}

			return appendErrReferences(err, errRefsManage)
		}

		if err = managed.validate(); err != nil {
			return appendErrReferences(err, errRefsManage)
		}

		for tkIdx := range managed.Tokens {
			tkn := managed.Tokens[tkIdx]
			num := tkIdx + 1
			//nolint
			errRefTkn := append(errRefsManage, fmt.Sprintf("access_token seq num: %d (name: %s)", num, tkn.Name))
			if err = tkn.validate(); err != nil {
				return appendErrReferences(err, errRefTkn)
			}

			for hkIdx := range managed.Tokens[tkIdx].Hooks {
				hook := managed.Tokens[tkIdx].Hooks[hkIdx]
				//nolint
				errRefsHook := append(errRefTkn, fmt.Sprintf("hook seq num: %d", hkIdx+1))

				if err = hook.validate(); err != nil {
					return appendErrReferences(err, errRefsHook)
				}

				// use_token hook validations
				if hook.Type == HookTypeUseToken {
					if managed.Type != ManagedTypePersonal {
						return appendErrReferences(ErrValidationHookUseTokenNotByPersonalType, errRefsHook)
					}
					if hookUseTokenUsed {
						return appendErrReferences(ErrValidationHookUseTokenAlreadyUse, errRefsHook)
					}

					if hkIdx != 0 {
						return appendErrReferences(ErrValidationHookUseTokenNotFirstSeq, errRefsHook)
					}

					hookUseTokenUsed = true
				}
			}
		}
	}

	return nil
}

// InitValues filling out the values
func (c *Config) InitValues() error {
	// evaluate contents from environment variable
	c.Token = evalEnvVar(c.Token)
	c.Host = evalEnvVar(c.Host)

	for idx := range c.Managed {
		managed := c.Managed[idx]
		for tkIdx := range managed.Tokens {
			tkn := managed.Tokens[tkIdx]

			if tkn.RenewBefore == "" {
				c.Managed[idx].Tokens[tkIdx].RenewBefore = c.DefaultRenewBefore
			}

			if tkn.ExpiryAfterRotate == "" {
				c.Managed[idx].Tokens[tkIdx].ExpiryAfterRotate = c.DefaultExpiryAfterRotate
			}

			for hkIdx := range tkn.Hooks {
				if tkn.Hooks[hkIdx].Retry == 0 {
					c.Managed[idx].Tokens[tkIdx].Hooks[hkIdx].Retry = c.DefaultHookRetry
				}
			}
		}
	}

	return c.validate()
}

// NewConfig initiate configuration with default values
func NewConfig() *Config {
	// setup default values
	cfg := Config{
		Host:                     defaultHost,
		Token:                    defaultToken,
		DefaultHookRetry:         defaultHookRetry,
		DefaultRenewBefore:       defaultRenewBefore,
		DefaultExpiryAfterRotate: defaultExpiryAfterRotate,
	}

	return &cfg
}
