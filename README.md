# blogGator

Add RSS feeds from across the internet to be collected
Store the collected posts in a PostgreSQL database
Follow and unfollow RSS feeds that other users have added
View summaries of the aggregated posts in the terminal, with a link to the full post

## Prerequisites

Need to install Go and PostgreSQL

Install gator using go install

## Usage

To set up to the config file, create a file named .gatorconfig.json in the home directory.
Example:

`{"db_url":"postgres://postgres:postgres@localhost:5432/gator?sslmode=disable","current_user_name":""}`

Some example commands

### Register User:

`gator register <user>`

### Login:

`gator login <user>`

### Add Feed:

`gator addfeed <url>`

### Follow feed:

`gator follow <url>`

### Aggregate posts from feeds into database:

`gator agg <duration>`

#### Avoid sending too many requests to the servers. Use Ctrl-c to interrupt the aggregation function

### Show posts from database from followed feeds:

`gator browse [optional: limit]`

