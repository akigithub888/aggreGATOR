package main

import (
	"context"
	"fmt"

	"github.com/akigithub888/aggreGATOR/internal/database"
)

func middlewareLoggedIn(handler func(s *state, cmd command, user database.GetUserByNameRow) error) func(*state, command) error {
	return func(s *state, cmd command) error {

		ctx := context.Background()

		if s.cfg.CurrentUserName == "" {
			return fmt.Errorf("you must be logged in to run this command")
		}

		user, err := s.db.GetUserByName(ctx, s.cfg.CurrentUserName)
		if err != nil {
			return fmt.Errorf("current user not found")
		}

		return handler(s, cmd, user)
	}
}
