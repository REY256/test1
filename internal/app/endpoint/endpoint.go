package endpoint

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"test1/graph/model"
	"test1/internal/app/service"
)

type Endpoint struct {
	s *service.Service
}

func New(s *service.Service) *Endpoint {
	return &Endpoint{s}
}

func (e *Endpoint) AddUserHandler(w http.ResponseWriter, r *http.Request) {
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

	user, err := e.s.AddUser(ctx, u)
	if err != nil {
		w.Write([]byte("add user error"))
		return
	}

	userBytes, err := json.Marshal(user)
	if err != nil {
		log.Fatal(err)
	}

	w.Write(userBytes)
}

func (e *Endpoint) GetUserByIdHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	i := r.URL.Query().Get("id")
	id, err := strconv.Atoi(i)
	if err != nil {
		w.Write([]byte("invalid id"))
		return
	}

	res, err := e.s.GetUserById(ctx, id)
	if err != nil {
		w.Write([]byte("error"))
		return
	}

	userBytes, err := json.Marshal(res)
	if err != nil {
		w.Write([]byte("invalid id"))
		return
	}

	w.Write(userBytes)
}

func (e *Endpoint) GetUsersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	res, err := e.s.GetUsers(ctx)
	if err != nil {
		w.Write([]byte("error"))
		return
	}

	userBytes, err := json.Marshal(res)
	if err != nil {
		w.Write([]byte("invalid id"))
		return
	}

	w.Write(userBytes)
}

func (e *Endpoint) ChangeUserHandler(w http.ResponseWriter, r *http.Request) {
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

	e.s.ChangeUserById(ctx, u)

	w.Write(body)
}

func (e *Endpoint) DeleteUserByIdHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	i := r.URL.Query().Get("id")
	id, err := strconv.Atoi(i)
	if err != nil {
		w.Write([]byte("invalid id"))
		return
	}

	user, err := e.s.DeleteUserById(ctx, id)
	if err != nil {
		w.Write([]byte("delete user error"))
		return
	}

	userBytes, err := json.Marshal(user)
	if err != nil {
		w.Write([]byte("error marshal user"))
		return
	}

	w.Write(userBytes)
}
