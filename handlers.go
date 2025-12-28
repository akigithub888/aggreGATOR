package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/akigithub888/aggreGATOR/internal/config"
	"github.com/akigithub888/aggreGATOR/internal/database"
	"github.com/google/uuid"
)

type state struct {
	db  *database.Queries
	cfg *config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	handlers map[string]func(*state, command) error
}

func handlerFollowing(s *state, cmd command) error {
	ctx := context.Background()

	user, err := s.db.GetUserByName(ctx, s.cfg.CurrentUserName)
	if err != nil {
		return fmt.Errorf("current user not found")
	}

	follows, err := s.db.GetFeedFollowsForUser(ctx, user.ID)
	if err != nil {
		return err
	}

	if len(follows) == 0 {
		fmt.Println("You are not following any feeds.")
		return nil
	}

	fmt.Println("You are following:")
	for _, follow := range follows {
		fmt.Println("-", follow.FeedName)
	}

	return nil
}

func handlerFollow(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("usage: follow <feed_url>")
	}
	ctx := context.Background()
	feedURL := cmd.args[0]

	user, err := s.db.GetUserByName(ctx, s.cfg.CurrentUserName)
	if err != nil {
		return fmt.Errorf("current user not found")
	}

	feed, err := s.db.GetFeedByURL(ctx, feedURL)
	if err != nil {
		return fmt.Errorf("feed not found for url %s", feedURL)
	}

	feedFollow, err := s.db.CreateFeedFollow(
		ctx,
		database.CreateFeedFollowParams{
			ID:        uuid.New(),
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
			UserID:    user.ID,
			FeedID:    feed.ID,
		})
	if err != nil {
		return err
	}
	fmt.Println("Feed follow created:")
	fmt.Println("ID:", feedFollow.ID)
	fmt.Println("User:", feedFollow.UserName)
	fmt.Println("Feed:", feedFollow.FeedName)

	return nil
}

func handlerFeeds(s *state, cmd command) error {
	if len(cmd.args) != 0 {
		return fmt.Errorf("Feeds does not take any arguments")
	}
	ctx := context.Background()

	feeds, err := s.db.GetFeeds(ctx)
	if err != nil {
		return fmt.Errorf("failed to get feeds: %w", err)
	}
	for _, feed := range feeds {
		fmt.Println("Feed:")
		fmt.Printf("  Name: %s\n", feed.FeedName)
		fmt.Printf("  URL: %s\n", feed.FeedUrl)
		fmt.Printf("  Created by: %s\n", feed.UserName)
		fmt.Println()
	}
	return nil
}

func handlerAddFeed(s *state, cmd command) error {
	if len(cmd.args) != 2 {
		return fmt.Errorf("usage: addfeed <name> <url>")
	}

	name := cmd.args[0]
	url := cmd.args[1]

	ctx := context.Background()

	user, err := s.db.GetUserByName(ctx, s.cfg.CurrentUserName)
	if err != nil {
		return fmt.Errorf("failed to get current user: %w", err)
	}

	feed, err := s.db.CreateFeed(ctx, database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
		Url:       url,
		UserID:    user.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to create feed: %w", err)
	}
	user, err = s.db.GetUserByName(ctx, s.cfg.CurrentUserName)
	if err != nil {
		return fmt.Errorf("current user not found")
	}
	_, err = s.db.CreateFeedFollow(ctx, database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	})
	if err != nil {
		return err
	}
	fmt.Println("Feed added and followed:")
	fmt.Println("Feed:", feed.Name)

	return nil
}

func handlerAgg(s *state, cmd command) error {
	if len(cmd.args) != 0 {
		return fmt.Errorf("Agg does not take any arguments")
	}
	ctx := context.Background()
	url := "https://www.wagslane.dev/index.xml"
	feed, err := fetchFeed(ctx, url)
	if err != nil {
		return fmt.Errorf("unable to fetchFeed: %w", err)
	}
	b, _ := json.MarshalIndent(feed, "", " ")
	fmt.Println(string(b))
	return nil
}

func handlerGetUsers(s *state, cmd command) error {
	if len(cmd.args) != 0 {
		return fmt.Errorf("users does not take any arguments")
	}
	ctx := context.Background()
	users, err := s.db.GetUsers(ctx)
	if err != nil {
		return fmt.Errorf("failed to get all users: %w", err)
	}
	for _, user := range users {
		if s.cfg.CurrentUserName == user.Name {
			fmt.Printf("* %s (current)\n", user.Name)
		} else {
			fmt.Printf("* %s\n", user.Name)
		}
	}
	return nil
}

func handlerReset(s *state, cmd command) error {
	if len(cmd.args) != 0 {
		return fmt.Errorf("reset does not take any arguments")
	}
	ctx := context.Background()
	err := s.db.DeleteAllUsers(ctx)
	if err != nil {
		return fmt.Errorf("failed to reset users: %w", err)
	}
	s.cfg.CurrentUserName = ""
	if err := config.Write(*s.cfg); err != nil {
		return fmt.Errorf("failed to clear config: %w", err)
	}
	fmt.Println("All users have been deleted.")
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("username is required")
	}
	username := cmd.args[0]
	ctx := context.Background()
	_, err := s.db.GetUserByName(ctx, username)
	if err == nil {
		// User exists
		fmt.Println("Username already exists.")
		os.Exit(1)
	} else if err != sql.ErrNoRows {
		// Some other error occurred
		return fmt.Errorf("failed to check username: %w", err)
	}
	params := database.CreateUserParams{
		ID:        uuid.New(),
		Name:      username,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	user, err := s.db.CreateUser(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	fmt.Println("DEBUG: User data:")
	fmt.Printf("  ID:        %s\n", user.ID)
	fmt.Printf("  Name:      %s\n", user.Name)
	fmt.Printf("  CreatedAt: %s\n", user.CreatedAt)
	fmt.Printf("  UpdatedAt: %s\n", user.UpdatedAt)
	s.cfg.CurrentUserName = username
	if err := config.Write(*s.cfg); err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}
	fmt.Printf("Current user set to: %s\n", username)
	return nil
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("username is required")
	}
	username := cmd.args[0]
	ctx := context.Background()
	user, err := s.db.GetUserByName(ctx, username)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("Error: user not registered.")
			os.Exit(1)
		}
		return fmt.Errorf("failed to check username: %w", err)
	}
	s.cfg.CurrentUserName = user.Name
	if err = config.Write(*s.cfg); err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	fmt.Println("User set to:", username)
	return nil
}

func (c *commands) run(s *state, cmd command) error {
	handler, ok := c.handlers[cmd.name]
	if !ok {
		return fmt.Errorf("unknown command: %s", cmd.name)
	}
	return handler(s, cmd)
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.handlers[name] = f
}
