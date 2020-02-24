package main

import (
	"errors"

	joehttp "github.com/go-joe/http-server"
	"github.com/go-joe/joe"
	"github.com/go-joe/redis-memory"
	"github.com/go-joe/slack-adapter"
)

// Config holds all parameters to setup a new chat bot.
type Config struct {
	SlackToken string // slack token, if empty we fallback to the CLI
	HTTPListen string // optional HTTP listen address to receive callbacks
	RedisAddr  string // optional address to store keys in Redis
}

// Modules creates a list of joe.Modules that can be used with this configuration.
func (conf Config) Modules() []joe.Module {
	var modules []joe.Module

	if conf.SlackToken != "" {
		modules = append(modules, slack.Adapter(conf.SlackToken))
	}

	if conf.HTTPListen != "" {
		modules = append(modules, joehttp.Server(conf.HTTPListen))
	}

	if conf.RedisAddr != "" {
		modules = append(modules, redis.Memory(conf.RedisAddr))
	}

	return modules
}

func (conf Config) Validate() error {
	if conf.HTTPListen == "" {
		return errors.New("missing HTTP listen address")
	}
	return nil
}
