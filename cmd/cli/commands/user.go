package commands

import (
	"context"

	"github.com/jonesrussell/goforms/internal/domain/user"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	userStore "github.com/jonesrussell/goforms/internal/infrastructure/persistence/store/user"
	"github.com/urfave/cli/v2"
)

func CreateUser(c *cli.Context) error {
	ctx := context.Background()

	// Initialize database connection
	db, err := getDB()
	if err != nil {
		return err
	}
	defer db.Close()

	// Create logger
	logger, err := logging.NewFactory().CreateLogger()
	if err != nil {
		return err
	}

	// Create user store
	store := userStore.NewStore(db, logger)

	// Create user service
	userService := user.NewService(store, logger, "your-jwt-secret")

	// Create new user
	newUser := &user.User{
		Email:     c.String("email"),
		FirstName: c.String("first-name"),
		LastName:  c.String("last-name"),
		Role:      c.String("role"),
		Active:    true,
	}

	// Set password
	if err := newUser.SetPassword(c.String("password")); err != nil {
		return err
	}

	// Save user
	createdUser, err := userService.SignUp(ctx, &user.Signup{
		Email:     newUser.Email,
		Password:  c.String("password"),
		FirstName: newUser.FirstName,
		LastName:  newUser.LastName,
	})
	if err != nil {
		return err
	}

	logger.Info("Successfully created user", logging.UintField("id", createdUser.ID))
	return nil
}

func ListUsers(c *cli.Context) error {
	ctx := context.Background()

	// Initialize database connection
	db, err := getDB()
	if err != nil {
		return err
	}
	defer db.Close()

	// Create logger
	logger, err := logging.NewFactory().CreateLogger()
	if err != nil {
		return err
	}

	// Create user store
	store := userStore.NewStore(db, logger)

	// Create user service
	userService := user.NewService(store, logger, "your-jwt-secret")

	// Get all users
	users, err := userService.ListUsers(ctx)
	if err != nil {
		return err
	}

	// Print users
	logger.Info("Users:")
	logger.Info("ID\tEmail\t\tName\t\tRole\tActive")
	logger.Info("--\t-----\t\t----\t\t----\t------")
	for i := range users {
		u := &users[i]
		logger.Info("User details",
			logging.UintField("id", u.ID),
			logging.StringField("email", u.Email),
			logging.StringField("name", u.FirstName+" "+u.LastName),
			logging.StringField("role", u.Role),
			logging.BoolField("active", u.Active),
		)
	}
	return nil
}

func DeleteUser(c *cli.Context) error {
	ctx := context.Background()

	// Initialize database connection
	db, err := getDB()
	if err != nil {
		return err
	}
	defer db.Close()

	// Create logger
	logger, err := logging.NewFactory().CreateLogger()
	if err != nil {
		return err
	}

	// Create user store
	store := userStore.NewStore(db, logger)

	// Create user service
	userService := user.NewService(store, logger, "your-jwt-secret")

	// Delete user
	userID := c.Uint("id")
	if err := userService.DeleteUser(ctx, userID); err != nil {
		return err
	}

	logger.Info("Successfully deleted user", logging.UintField("id", userID))
	return nil
}
