package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	user "github.com/KostyShatovGO/tasksync/internal"
	_ "github.com/jackc/pgx/v5/stdlib"
	"golang.org/x/crypto/bcrypt"
)

var DB *sql.DB

func InitDB() error {
	databaseUrl := os.Getenv("DATABASE_URL")
	if databaseUrl == "" {
		return fmt.Errorf("DATABASE_URL not set in .env")
	}

	var err error
	DB, err = sql.Open("postgres", databaseUrl)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	err = DB.Ping()
	DB.SetMaxOpenConns(10)
	DB.SetMaxIdleConns(5)
	if err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}
	log.Println("Successfully connected to PostgreSQL!")
	return nil
}

func CreateUser(username, plainPassword string) (*user.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	query := `
		INSERT INTO users (username,password)
		VALUES ($1, $2)
		RETURNING id, username, created_at
	`
	var newUser user.User
	err = DB.QueryRowContext(ctx, query, username, string(hashedPassword)).Scan(
		&newUser.ID,
		&newUser.Username,
		&newUser.CreatedAt, // Теперь это поле существует
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}
	log.Printf("User %s created with ID %d", username, newUser.ID)

	return &newUser, nil
}

func GetUserByUsername(username string) (*user.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	query := `
		SELECT id, username, password, created_at
		FROM users
		WHERE username = $1
	`
	var foundUser user.User
	err := DB.QueryRowContext(ctx, query, username).Scan(
		&foundUser.ID,
		&foundUser.Username,
		&foundUser.Password,
		&foundUser.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %v", err)
	}
	log.Printf("User %s found with ID %d", username, foundUser.ID)
	return &foundUser, nil
}
func CloseDB() {
	if DB != nil {
		DB.Close()
	}
}
