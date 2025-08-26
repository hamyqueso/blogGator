package main

import (
	"errors"

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
	}

	username := cmd.args[0]

	s.cfg.CurrentUserName = username

	return nil
}

type commands struct {
	handlers map[string]func(*state, command) error
}
