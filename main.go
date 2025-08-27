package main

import (
	"fmt"
	"os"

	"github.com/hamyqueso/blogGator/internal/config"
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

	s := &state{cfg: &c}
	if len(os.Args) < 2 {
		fmt.Println("Need at least 2 arguments")
		os.Exit(1)
	}
	cmd := command{name: os.Args[1], args: os.Args[2:]}
	cmds := commands{handlers: make(map[string]func(*state, command) error)}

	cmds.register("login", handlerLogin)

	err = cmds.run(s, cmd)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	c, err = config.Read()
	if err != nil {
		fmt.Printf("%v\n", err)
	}

	fmt.Println(c.DBURL)
	fmt.Println(c.CurrentUserName)
}
