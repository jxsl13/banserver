package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

var (
	errAddressPasswordMismatch = errors.New("the number of ECON_PASSWORD doesn't match the number of ECON_ADDRESSES, either provide one password for all addresses or one password per address")
)

// New creates a new configuration file based on
// the data that has been retrieved from the .env environment file.
// any call after the first one will return the config of the first call
// the location of the .env file can be changed via the DefaultEnvFile variable
func New() *Config {
	return &Config{
		EconReconnectDelay:   10 * time.Second,
		EconReconnectTimeout: 24 * time.Hour,
		PermaBanReason:       "permanently banned",
		PermaBanDuration:     24 * time.Hour,
		ChatBanReason:        "prohibited chat message",
		ChatBanDuration:      24 * time.Hour,
	}
}

// Config represents the application configuration
type Config struct {
	EconServersString string `koanf:"econ.addresses" validate:"required" description:"comma separated list of econ addresses (<ip/hostname>:port)"`
	EconServers       []string

	EconPasswordsString  string `koanf:"econ.passwords" validate:"required" description:"comma separated list of econ passwords"`
	EconPasswords        []string
	EconReconnectDelay   time.Duration `koanf:"econ.reconnect.delay" validate:"required"`
	EconReconnectTimeout time.Duration `koanf:"econ.reconnect.timeout" validate:"required"`

	IPBlacklistsString  string `koanf:"ip.blacklists" description:"comma separated list of files containing ip ranges to blacklist"`
	IPBlacklists        []string
	ChatBlacklistString string `koanf:"chat.blacklists" description:"comma separated list that contains regular expressions to check message blacklists"`

	ChatBlacklists []string

	Propagate bool `koanf:"propagate" description:"propagate bans and unbans from one game server to all other game servers"`

	PermaBanReason   string        `koanf:"perma.ban.reason" description:"default reason for permabans"`
	PermaBanDuration time.Duration `koanf:"perma.ban.duration" description:"default duration for permabans"`

	ChatBanReason   string        `koanf:"chat.ban.reason" description:"default reason for chat bans"`
	ChatBanDuration time.Duration `koanf:"chat.ban.duration" description:"default duration for chat bans"`
}

func (c *Config) Validate() error {
	err := validator.New().Struct(c)
	if err != nil {
		return err
	}

	if c.PermaBanDuration < time.Minute {
		return errors.New("perma ban duration must be at least 1m")
	}

	if len(c.PermaBanReason) == 0 {
		return errors.New("perma ban reason must not be empty")
	}

	if c.ChatBanDuration < time.Minute {
		return errors.New("chat ban duration must be at least 1m")
	}

	if len(c.ChatBanReason) == 0 {
		return errors.New("chat ban reason must not be empty")
	}

	if len(c.EconServersString) == 0 {
		return errors.New("econ addresses must not be empty")
	}

	c.EconServers = strings.Split(c.EconServersString, ",")
	c.EconPasswords = strings.Split(c.EconPasswordsString, ",")

	// add password for every econ server.
	if len(c.EconServers) != len(c.EconPasswords) {
		if len(c.EconServers) > 1 && len(c.EconPasswords) > 1 {
			return errAddressPasswordMismatch
		}
		if len(c.EconServers) > 1 && len(c.EconPasswords) == 1 {
			for len(c.EconPasswords) < len(c.EconServers) {
				c.EconPasswords = append(c.EconPasswords, c.EconPasswords[0])
			}
		}
	}

	if len(c.IPBlacklistsString) > 0 {
		c.IPBlacklists = strings.Split(c.IPBlacklistsString, ",")

		// check if all files exist
		for _, file := range c.IPBlacklists {
			if err := fileMustExist(file); err != nil {
				return fmt.Errorf("ip blacklist file %s does not exist: %w", file, err)
			}
		}
	}

	if len(c.ChatBlacklistString) > 0 {
		c.ChatBlacklists = strings.Split(c.ChatBlacklistString, ",")

		// check if all files exist
		for _, file := range c.ChatBlacklists {
			if err := fileMustExist(file); err != nil {
				return fmt.Errorf("chat blacklist file %s does not exist: %w", file, err)
			}
		}
	}

	if !c.Propagate && len(c.ChatBlacklists) == 0 && len(c.IPBlacklists) == 0 {
		return fmt.Errorf("pointless configuration, you need to have at least propagate bans enabled or chat blacklist or ip blacklist defined")
	} else if len(c.ChatBlacklists) == 0 && len(c.IPBlacklists) == 0 && c.Propagate && len(c.EconServers) < 2 {
		return fmt.Errorf("pointless configuration, you need to have at least two game servers (= econ addresses) to propagate bans")
	}

	return nil
}

func fileMustExist(path string) error {
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}

	if !fi.Mode().IsRegular() {
		return fmt.Errorf("file %s is not a regular file", path)
	}
	return nil
}
