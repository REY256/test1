package graph

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"test1/graph/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	Pool *pgxpool.Pool
}

func (app *Resolver) AddUser(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.Write([]byte("read body err"))
		return
	}

	user := model.User{}
	if err := json.Unmarshal(body, &user); err != nil {
		w.Write([]byte("invalid user"))
		return
	}

	row := app.Pool.QueryRow(ctx, "insert into test_table (name, surname, patronymic, age, gender) values($1, $2, $3, $4, $5) returning id")
	row.Scan(&user.ID)

	userBytes, err := json.Marshal(user)
	if err != nil {
		log.Fatal(err)
	}

	w.Write(userBytes)
}

func (app *Resolver) GetUserById(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	i := r.URL.Query().Get("id")
	id, err := strconv.Atoi(i)
	if err != nil {
		w.Write([]byte("invalid id"))
		return
	}

	var u model.User
	row := app.Pool.QueryRow(ctx, "select id, name, surname, patronymic, age, gender from test_table where id = $1", id)
	row.Scan(&u.ID, &u.Name, &u.Surname, &u.Patronymic, u.Age, u.Gender)

	userBytes, err := json.Marshal(u)
	if err != nil {
		w.Write([]byte("error"))
	}

	w.Write(userBytes)
}

func (app *Resolver) ChangeUser(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.Write([]byte("read body err"))
		return
	}

	u := model.User{}
	if err := json.Unmarshal(body, &u); err != nil {
		w.Write([]byte("invalid user"))
		return
	}

	app.Pool.Exec(ctx, "update test_table set name = $1, surname = $2, patronymic = $3, age = $4, gender = $5) where id = $6", u.Name, u.Surname, u.Patronymic, u.Age, u.Gender, u.ID)

	w.Write(body)
}
