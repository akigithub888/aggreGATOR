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
