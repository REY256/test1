package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"test1/graph/model"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	Pool *pgxpool.Pool
	Rdb  *redis.Client
}

func (app *Resolver) AddUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.Write([]byte("read body err"))
		return
	}

	var u model.NewUser
	if err := json.Unmarshal(body, &u); err != nil {
		w.Write([]byte("invalid user"))
		return
	}

	user := app.AddUser(ctx, u)

	userBytes, err := json.Marshal(user)
	if err != nil {
		log.Fatal(err)
	}

	w.Write(userBytes)
}

func (app *Resolver) AddUser(ctx context.Context, u model.NewUser) model.User {
	var id int

	app.Pool.QueryRow(ctx, "insert into test_table (name, surname, patronymic, age, gender) values($1, $2, $3, $4, $5) returning id", u.Name, u.Surname, u.Patronymic, u.Age, u.Gender).Scan(&id)

	user := model.User{
		ID:         id,
		Name:       u.Name,
		Surname:    u.Surname,
		Patronymic: u.Patronymic,
		Age:        u.Age,
		Gender:     u.Gender,
		Country:    []*model.Country{},
	}

	return user
}

func (app *Resolver) GetUserByIdHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	i := r.URL.Query().Get("id")
	id, err := strconv.Atoi(i)
	if err != nil {
		w.Write([]byte("invalid id"))
		return
	}

	res, err := app.GetUserById(ctx, id)
	if err != nil {
		w.Write([]byte("error"))
		return
	}

	w.Write(res)
}

func (app *Resolver) GetUserById(ctx context.Context, id int) ([]byte, error) {
	res, err := app.Rdb.Get(ctx, fmt.Sprint(id)).Result()
	if err != nil {
		var u model.User

		row := app.Pool.QueryRow(ctx, "select id, name, surname, patronymic, age, gender from test_table where id = $1", id)
		row.Scan(&u.ID, &u.Name, &u.Surname, &u.Patronymic, u.Age, u.Gender)

		userBytes, err := json.Marshal(u)
		if err != nil {
			return nil, err
		}

		app.Rdb.Set(ctx, fmt.Sprint(id), userBytes, 0)

		return userBytes, nil
	}

	return []byte(res), nil
}

func (app *Resolver) ChangeUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.Write([]byte("read body err"))
		return
	}

	u := model.ChangeUser{}
	if err := json.Unmarshal(body, &u); err != nil {
		w.Write([]byte("invalid user"))
		return
	}

	app.ChangeUserById(ctx, u)

	w.Write(body)
}

func (app *Resolver) ChangeUserById(ctx context.Context, u model.ChangeUser) {
	app.Pool.Exec(ctx, "update test_table set name = $1, surname = $2, patronymic = $3, age = $4, gender = $5) where id = $6", u.Name, u.Surname, u.Patronymic, u.Age, u.Gender, u.ID)
}

func (app *Resolver) DeleteUserByIdHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	i := r.URL.Query().Get("id")
	id, err := strconv.Atoi(i)
	if err != nil {
		w.Write([]byte("invalid id"))
		return
	}

	u := app.DeleteUserById(ctx, id)

	userBytes, err := json.Marshal(u)
	if err != nil {
		w.Write([]byte("error"))
		return
	}

	w.Write(userBytes)
}

func (app *Resolver) DeleteUserById(ctx context.Context, id int) model.User {
	var u model.User

	row := app.Pool.QueryRow(ctx, "delete from test_table where id = $1 returnint id, name, surname, patronymic, age, gender", id)
	row.Scan(&u.ID, &u.Name, &u.Surname, &u.Patronymic, u.Age, u.Gender)

	return u
}
