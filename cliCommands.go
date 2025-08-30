package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/hamyqueso/blogGator/internal/config"
	"github.com/hamyqueso/blogGator/internal/database"
	"github.com/lib/pq"
)

type state struct {
	cfg *config.Config
	db  *database.Queries
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

	u, err := s.db.GetUser(context.Background(), username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			fmt.Println("user not found")
			os.Exit(1)
		}
		return err
	}

	if err := s.cfg.SetUser(u.Name); err != nil {
		return err
	}

	fmt.Println("Username has been set in config file")

	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return errors.New("register command requires a name argument")
	}

	params := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.args[0],
	}

	u, err := s.db.CreateUser(context.Background(), params)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			fmt.Println("user was already registered")
			os.Exit(1)
		}
		return err
	}

	s.cfg.SetUser(u.Name)

	fmt.Printf("created user %s\n", u.Name)
	log.Printf("user: %+v\n", u)
	return nil
}

func handlerReset(s *state, cmd command) error {
	err := s.db.Reset(context.Background())
	if err != nil {
		fmt.Println("error resetting database")
		return err
	}

	fmt.Println("database reset successful")
	return nil
}

func handlerUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		fmt.Println("error getting users")
		return err
	}

	currentUser := s.cfg.CurrentUserName

	for _, user := range users {
		result := "* " + user
		if user == currentUser {
			result += " (current)"
		}

		fmt.Println(result)
	}

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
