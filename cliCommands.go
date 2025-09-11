package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"html"
	"log"
	"os"
	"strconv"
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
	// url := "https://www.wagslane.dev/index.xml"
	//
	// feed, err := fetchFeed(context.Background(), url)
	// if err != nil {
	// 	fmt.Println("error fetching rss feed")
	// 	return err
	// }
	//
	// fmt.Println(feed)

	if len(cmd.args) < 1 {
		return errors.New("agg command requires time duration in seconds")
	}

	duration, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return err
	}

	fmt.Printf("Collecting feeds every %v\n", duration)

	ticker := time.NewTicker(duration)
	for ; ; <-ticker.C {
		// fmt.Println("ticking...")
		scrapeFeeds(s)
	}

	return nil
}

func scrapeFeeds(s *state) {
	feed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		fmt.Printf("%v\n", err)
	}

	fmt.Printf("working on feed: %s\n", feed.Name)

	err = s.db.MarkFeedFetched(context.Background(), feed.ID)
	if err != nil {
		fmt.Printf("%v\n", err)
	}

	// fmt.Printf("feed marked. last fetched at: %s", feed.LastFetchedAt.Time)

	rss, err := fetchFeed(context.Background(), feed.Url)
	if err != nil {
		fmt.Printf("%v\n", err)
	}

	for _, item := range rss.Channel.Item {
		// if item.Title != "" {
		// 	fmt.Printf("* %s\n", item.Title)
		// 	fmt.Printf("%s\n", item.PubDate)
		// }
		//
		const layout = "Mon, 02 Jan 2006 03:04:05 +0700"
		t, err := time.Parse(time.RFC1123Z, item.PubDate)
		if err != nil {
			fmt.Printf("%v\n", err)
		}

		var description sql.NullString

		if item.Description != "" {
			description = sql.NullString{
				String: html.UnescapeString(item.Description),
				Valid:  true,
			}
		} else {
			description = sql.NullString{
				String: "",
				Valid:  false,
			}
		}
		params := database.CreatePostParams{
			Title:       item.Title,
			Url:         item.Link,
			Description: description,
			PublishedAt: t,
			FeedID:      feed.ID,
		}

		var pqErr *pq.Error

		_, err = s.db.CreatePost(context.Background(), params)
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		} else if err != nil {
			fmt.Printf("%v\n", err)
		}
	}
}

func handlerBrowse(s *state, cmd command, user database.User) error {
	var limit int64
	var err error
	if len(cmd.args) < 1 {
		limit = 2
	} else {

		limit, err = strconv.ParseInt(cmd.args[0], 10, 32)
		if err != nil {
			return err
		}
	}

	params := database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  int32(limit),
	}

	posts, err := s.db.GetPostsForUser(context.Background(), params)
	if err != nil {
		return err
	}

	for _, post := range posts {
		fmt.Println()
		fmt.Printf("Title: %v\n", post.Title)
		fmt.Printf("Blog: %v\n", post.BlogName)
		if post.Description.Valid {
			fmt.Printf("Description: %v\n", post.Description.String)
		}
		fmt.Printf("Link: %v\n", post.Url)
		fmt.Printf("Timestamp: %v\n", post.DisplayTime)

	}

	return nil
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 2 {
		return errors.New("error add feed command requires two parameters (name, url)")
	}

	// user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
	// if err != nil {
	// 	return err
	// }

	params := database.CreateFeedParams{
		Name:   cmd.args[0],
		Url:    cmd.args[1],
		UserID: user.ID,
	}

	feed, err := s.db.CreateFeed(context.Background(), params)
	if err != nil {
		return err
	}

	createFollowParams := database.CreateFeedFollowParams{
		FeedID: feed.ID,
		UserID: user.ID,
	}

	_, err = s.db.CreateFeedFollow(context.Background(), createFollowParams)
	if err != nil {
		return nil
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

func handlerFeedFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 1 {
		return errors.New("follow command requires a url argument")
	}
	// user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
	// if err != nil {
	// 	return err
	// }

	feed, err := s.db.GetFeedByURL(context.Background(), cmd.args[0])
	if err != nil {
		return err
	}

	params := database.CreateFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	}

	_, err = s.db.CreateFeedFollow(context.Background(), params)
	if err != nil {
		return err
	}

	fmt.Printf("Feed name: %s; currrent user: %s\n", feed.Name, user.Name)

	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	feeds, err := s.db.GetFollowingFeeds(context.Background(), user.ID)
	if err != nil {
		return err
	}

	fmt.Printf("Current User: %s\n", s.cfg.CurrentUserName)
	fmt.Println("Following:")
	for _, id := range feeds {
		feed, err := s.db.GetFeedByID(context.Background(), id)
		if err != nil {
			return err
		}
		fmt.Printf("* %s\n", feed.Name)
	}

	return nil
}

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		usr, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
		if err != nil {
			return err
		}
		err = handler(s, cmd, usr)
		if err != nil {
			return err
		}
		return nil
	}
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 1 {
		fmt.Println("unfollow command requires feed url parameter")
		os.Exit(1)
	}

	feed, err := s.db.GetFeedByURL(context.Background(), cmd.args[0])
	if err != nil {
		return err
	}
	params := database.UnfollowFeedParams{
		UserID: user.ID,
		FeedID: feed.ID,
	}
	err = s.db.UnfollowFeed(context.Background(), params)
	if err != nil {
		return err
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
