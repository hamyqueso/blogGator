package main

import (
	"fmt"

	"github.com/hamyqueso/blogGator/internal/config"
)

func main() {
	c, err := config.Read()
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	fmt.Println(c.DBURL)
	fmt.Println(c.CurrentUserName)
}
