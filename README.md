# Gator CLI

Gator is a simple RSS aggregator CLI written in Go. It allows you to manage RSS feeds, fetch posts, and browse them directly from the terminal.

## Requirements

Before using Gator, make sure you have the following installed:

- [Go](https://golang.org/dl/) (1.21+ recommended)
- [PostgreSQL](https://www.postgresql.org/download/) (latest stable version)

## Installation

To install the CLI globally on your system:

```bash
go install github.com/your-username/gator@latest
```

After installing, the `gator` command should be available in your terminal.

> Note: `go run .` is only for development; `gator` is what you run in production.

## Configuration

Gator requires a config file to connect to your database. Create a config file named `config.json` in the root of your project (next to `main.go` and `README.md`):

```json
{
    "db_url": "postgres://username:password@localhost:5432/gator",
    "current_user_name": ""
}
```

- Replace `username` and `password` with your PostgreSQL credentials.
- `current_user_name` will be automatically set when you log in.

Your project structure should look like this:

```
gator/
├── main.go
├── go.mod
├── go.sum
├── internal/
├── config.json
└── README.md
```

## Usage

### Development

```bash
go run . <command> [flags]
```

### Production

```bash
gator <command> [flags]
```

## Commands

- Register a new user:
```bash
gator register
```

- Login as an existing user:
```bash
gator login
```

- Add a new RSS feed (must be logged in):
```bash
gator addfeed https://xkcd.com/rss.xml XKCD
gator addfeed https://hnrss.org/frontpage "Hacker News"
```

- Fetch posts from all feeds:
```bash
gator agg
```

- Browse posts for the logged-in user:
```bash
gator browse --limit 5
```
> If `--limit` is omitted, the default is 2 posts.

- List all feeds:
```bash
gator feeds
```

- Follow another user:
```bash
gator follow username
```

- See the users you are following:
```bash
gator following
```

## Full Test Workflow

1. Register a new user:
```bash
go run . register
```

2. Login as that user:
```bash
go run . login
```

3. Add some RSS feeds:
```bash
go run . addfeed TechCrunch https://techcrunch.com/feed/
go run . addfeed HackerNews https://news.ycombinator.com/rss
go run . addfeed BootBlog https://blog.boot.dev/index.xml

```

4. Fetch posts from all feeds (must provide a time interval, e.g., "10s", "1m"):
```bash
go run . 10s
```

5. Browse posts:
```bash
go run . browse --limit 5
```

You should see the latest posts from your feeds, including titles, URLs, and published dates.

## Building the Binary

To compile a standalone binary:

```bash
go build -o gator
```

After building, you can run `./gator` without needing Go installed. On Unix systems, you can move it to `/usr/local/bin` to make it globally available:

```bash
sudo mv gator /usr/local/bin/
```

## Contributing

1. Fork the repository.
2. Make changes on your branch.
3. Push your branch and submit a pull request.

## GitHub Repo

Once pushed, submit the URL to your remote repository:

```
https://github.com/your-username/gator
```

> Go programs are statically compiled, so after building or installing, the CLI runs independently of the Go toolchain.
