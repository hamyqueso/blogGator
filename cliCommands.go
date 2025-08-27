package main

import (
	"errors"
	"fmt"

	"github.com/hamyqueso/blogGator/internal/config"
)

type state struct {
	cfg *config.Config
}

type command struct {
	name string
	args []string
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return errors.New("login command requires username argument")
	} else if len(cmd.args) > 1 {
		return errors.New("login command takes one username argument, too many arguments")
	}

	username := cmd.args[0]

	if err := s.cfg.SetUser(username); err != nil {
		return err
	}

	fmt.Println("Username has been set in config file")

	return nil
}

type commands struct {
	handlers map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	return c.handlers[cmd.name](s, cmd)
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.handlers[name] = f
}
