package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/bogdanrat/web-server/contracts/models"
	"github.com/bogdanrat/web-server/service/core/common"
	"github.com/bogdanrat/web-server/service/core/lib"
	"github.com/bogdanrat/web-server/service/core/store"
	_ "github.com/lib/pq"
	"strconv"
	"time"
)

const (
	getAllUsers         = `SELECT * FROM users;`
	getUserByEmail      = `SELECT * FROM users WHERE email = $1;`
	insertUser          = `INSERT INTO users (name, email, password, qr_secret) VALUES ($1, $2, $3, $4);`
	updateUserQRByEmail = `UPDATE users SET qr_secret = $2 WHERE email = $1`
)

type Repository struct {
	DB *sql.DB
}

func NewRepository() (store.DatabaseRepository, error) {
	secrets, err := lib.GetDatabaseSecrets()
	if err != nil {
		return nil, err
	}

	conn, err := initConnection(secrets)
	if err != nil {
		return nil, err
	}

	return &Repository{
		DB: conn,
	}, nil
}

func initConnection(secrets *common.DatabaseSecrets) (*sql.DB, error) {
	port := strconv.Itoa(secrets.Port)
	dataSource := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", secrets.Host, port, secrets.Username, secrets.Password, secrets.DbName, "disable")
	conn, err := sql.Open("postgres", dataSource)

	if err != nil {
		return nil, err
	}
	if err := conn.Ping(); err != nil {
		return nil, err
	}

	return conn, nil
}

func (repo *Repository) GetAllUsers() ([]*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	rows, err := repo.DB.QueryContext(ctx, getAllUsers)
	if err != nil {
		return nil, err
	}

	users := make([]*models.User, 0)

	for rows.Next() {
		user := &models.User{}

		err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.Password,
			&user.QRSecret,
		)
		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func (repo *Repository) GetUserByEmail(email string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	user := &models.User{}

	err := repo.DB.QueryRowContext(ctx, getUserByEmail, email).Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.QRSecret)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (repo *Repository) InsertUser(user *models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	_, err := repo.DB.ExecContext(ctx, insertUser, user.Name, user.Email, user.Password, user.QRSecret)
	if err != nil {
		return err
	}
	return nil
}

func (repo *Repository) UpdateUserQRSecret(email, secret string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	_, err := repo.DB.ExecContext(ctx, updateUserQRByEmail, email, secret)
	if err != nil {
		return err
	}
	return nil
}
