package postgres

import (
	"fmt"
	"github.com/Talodoak/todo-app/internal/models"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"strings"
)

type TodoListPostgres struct {
	db *sqlx.DB
}

func NewTodoListPostgres(db *sqlx.DB) *TodoListPostgres {
	return &TodoListPostgres{db: db}
}

func (r TodoListPostgres) Create(userId int, list models.TodoList) (int, error) {
	tx, err := r.db.Begin()
	if err != nil {
		logrus.Error("Create error: ", err)
		return 0, err
	}

	var id int
	createListQuery := fmt.Sprintf("INSERT INTO %s (title, description) VALUES ($1, $2) RETURNING id", viper.GetString("postgres.todoListsTable"))
	row := tx.QueryRow(createListQuery, list.Title, list.Description)
	if insertErr := row.Scan(&id); insertErr != nil {
		rlbckErr := tx.Rollback()
		if rlbckErr != nil {
			logrus.Error("createListQuery, error while rollback: ", rlbckErr)
			return 0, rlbckErr
		}
		logrus.Error("createListQuery, error while rollback: ", insertErr)
		return 0, insertErr
	}

	createUsersListQuery := fmt.Sprintf("INSERT INTO %s (user_id, list_id) VALUES ($1, $2)", viper.GetString("postgres.usersListsTable"))
	_, err = tx.Exec(createUsersListQuery, userId, id)
	if err != nil {
		listRlbckErr := tx.Rollback()
		if listRlbckErr != nil {
			logrus.Error("creatUsersListQuery, error while rollback: ", listRlbckErr)
			return 0, listRlbckErr
		}
		logrus.Error("createUsersListQuery, error while insert: ", err)
		return 0, err
	}

	return id, tx.Commit()
}

func (r *TodoListPostgres) GetAll(userId int) ([]models.TodoList, error) {
	var lists []models.TodoList

	query := fmt.Sprintf("SELECT tl.id, tl.title, tl.description FROM %s tl INNER JOIN %s ul on tl.id = ul.list_id WHERE ul.user_id = $1",
		viper.GetString("postgres.todoListsTable"), viper.GetString("postgres.usersListsTable"))
	err := r.db.Select(&lists, query, userId)
	logrus.Error("GetAll error: ", err)

	return lists, err
}

func (r *TodoListPostgres) GetById(userId, listId int) (models.TodoList, error) {
	var list models.TodoList

	query := fmt.Sprintf(`SELECT tl.id, tl.title, tl.description FROM %s tl
								INNER JOIN %s ul on tl.id = ul.list_id WHERE ul.user_id = $1 AND ul.list_id = $2`,
		viper.GetString("postgres.todoListsTable"), viper.GetString("postgres.usersListsTable"))
	err := r.db.Get(&list, query, userId, listId)
	logrus.Error("GetById error: ", err)

	return list, err
}

func (r *TodoListPostgres) Delete(userId, listId int) error {
	query := fmt.Sprintf("DELETE FROM %s tl USING %s ul WHERE tl.id = ul.list_id AND ul.user_id=$1 AND ul.list_id=$2",
		viper.GetString("postgres.todoListsTable"), viper.GetString("postgres.usersListsTable"))
	_, err := r.db.Exec(query, userId, listId)
	logrus.Error("Delete error: ", err)

	return err
}

func (r *TodoListPostgres) Update(userId, listId int, input models.UpdateListInput) error {
	setValues := make([]string, 0)
	args := make([]interface{}, 0)
	argId := 1

	if input.Title != nil {
		setValues = append(setValues, fmt.Sprintf("title=$%d", argId))
		args = append(args, *input.Title)
		argId++
	}

	if input.Description != nil {
		setValues = append(setValues, fmt.Sprintf("description=$%d", argId))
		args = append(args, *input.Description)
		argId++
	}
	setQuery := strings.Join(setValues, ", ")

	query := fmt.Sprintf("UPDATE %s tl SET %s FROM %s ul WHERE tl.id = ul.list_id AND ul.list_id=$%d AND ul.user_id=$%d",
		viper.GetString("postgres.todoListsTable"), setQuery, viper.GetString("postgres.usersListsTable"), argId, argId+1)
	args = append(args, listId, userId)

	logrus.Debugf("updateQuery: %s", query)
	logrus.Debugf("args: %s", args)

	_, err := r.db.Exec(query, args...)
	logrus.Error("Update error: ", err)
	return err
}
