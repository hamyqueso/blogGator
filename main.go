package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/hamyqueso/blogGator/internal/config"
	"github.com/hamyqueso/blogGator/internal/database"
	_ "github.com/lib/pq"
)

func main() {
	c, err := config.Read()
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	// if err = c.SetUser("jonny"); err != nil {
	// 	fmt.Printf("%v\n", err)
	// }

	db, err := sql.Open("postgres", c.DBURL)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	dbQueries := database.New(db)

	s := &state{cfg: &c, db: dbQueries}

	if len(os.Args) < 2 {
		fmt.Println("Need at least 2 arguments")
		os.Exit(1)
	}
	cmd := command{name: os.Args[1], args: os.Args[2:]}
	cmds := commands{handlers: make(map[string]func(*state, command) error)}

	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	cmds.register("users", handlerUsers)
	cmds.register("agg", handlerAgg)
	cmds.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	cmds.register("feeds", handlerFeeds)
	cmds.register("follow", middlewareLoggedIn(handlerFeedFollow))
	cmds.register("following", middlewareLoggedIn(handlerFollowing))
	cmds.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	cmds.register("browse", middlewareLoggedIn(handlerBrowse))

	err = cmds.run(s, cmd)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	c, err = config.Read()
	if err != nil {
		fmt.Printf("%v\n", err)
	}

	// fmt.Println(c.DBURL)
	// fmt.Println(c.CurrentUserName)
}
