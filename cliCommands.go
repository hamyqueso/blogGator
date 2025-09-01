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

func handlerAgg(s *state, cmd command) error {
	url := "https://www.wagslane.dev/index.xml"

	feed, err := fetchFeed(context.Background(), url)
	if err != nil {
		fmt.Println("error fetching rss feed")
		return err
	}

	fmt.Println(feed)
	return nil
}

func handlerAddFeed(s *state, cmd command) error {
	if len(cmd.args) < 2 {
		return errors.New("error add feed command requires two parameters (name, url)")
	}

	user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
	if err != nil {
		return err
	}

	params := database.CreateFeedParams{
		Name:   cmd.args[0],
		Url:    cmd.args[1],
		UserID: user.ID,
	}

	feed, err := s.db.CreateFeed(context.Background(), params)
	if err != nil {
		return err
	}

	fmt.Println(feed)
	return nil
}

func handlerFeeds(s *state, cmd command) error {
	items, err := s.db.ListFeeds(context.Background())
	if err != nil {
		return errors.New("error obtaining data from feeds table")
	}

	for _, item := range items {
		user, err := s.db.GetUserByID(context.Background(), item.UserID)
		if err != nil {
			return errors.New("error getting username from users table by id")
		}
		fmt.Printf("Feed name: %s; Feed URL: %s; Creating User: %s\n", item.Name, item.Url, user.Name)
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
