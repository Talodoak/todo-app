package postgres

import (
	"fmt"
	"github.com/Talodoak/todo-app/internal/models"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
)

type AuthPostgres struct {
	db *sqlx.DB
}

func NewAuthPostgres(db *sqlx.DB) *AuthPostgres {
	return &AuthPostgres{db: db}
}

func (r *AuthPostgres) CreateUser(user models.User) (int, error) {
	var id int

	query := fmt.Sprintf(`INSERT INTO %s (name, username, password_hash) values ($1, $2, $3) RETURNING id`, viper.GetString("postgres.usersTable"))
	row := r.db.QueryRow(query, user.Name, user.Username, user.Password)

	if err := row.Scan(&id); err != nil {
		return 0, err
	}

	return id, nil
}

func (r *AuthPostgres) GetUser(username, password string) (models.User, error) {
	var user models.User

	query := fmt.Sprintf(` 
	SELECT id 
	FROM %s 
	WHERE username=$1 
	    AND password_hash=$2
	`, viper.GetString("postgres.usersTable"))

	err := r.db.Get(&user, query, username, password)

	return user, err
}
