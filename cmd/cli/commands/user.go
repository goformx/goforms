package commands

import (
	"context"
	"fmt"

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

	// Create logger factory
	factory := logging.NewFactory()

	// Create logger
	logger, err := factory.CreateLogger()
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
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
		return fmt.Errorf("failed to set password: %w", err)
	}

	// Save user
	createdUser, err := userService.SignUp(ctx, &user.Signup{
		Email:     newUser.Email,
		Password:  c.String("password"),
		FirstName: newUser.FirstName,
		LastName:  newUser.LastName,
	})
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	fmt.Printf("Successfully created user with ID: %d\n", createdUser.ID)
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

	// Create logger factory
	factory := logging.NewFactory()

	// Create logger
	logger, err := factory.CreateLogger()
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}

	// Create user store
	store := userStore.NewStore(db, logger)

	// Create user service
	userService := user.NewService(store, logger, "your-jwt-secret")

	// Get all users
	users, err := userService.ListUsers(ctx)
	if err != nil {
		return fmt.Errorf("failed to list users: %w", err)
	}

	// Print users
	fmt.Println("\nUsers:")
	fmt.Println("ID\tEmail\t\tName\t\tRole\tActive")
	fmt.Println("--\t-----\t\t----\t\t----\t------")
	for _, u := range users {
		fmt.Printf("%d\t%s\t%s %s\t%s\t%v\n",
			u.ID, u.Email, u.FirstName, u.LastName, u.Role, u.Active)
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

	// Create logger factory
	factory := logging.NewFactory()

	// Create logger
	logger, err := factory.CreateLogger()
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}

	// Create user store
	store := userStore.NewStore(db, logger)

	// Create user service
	userService := user.NewService(store, logger, "your-jwt-secret")

	// Delete user
	userID := c.Uint("id")
	if err := userService.DeleteUser(ctx, userID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	fmt.Printf("Successfully deleted user with ID: %d\n", userID)
	return nil
}
