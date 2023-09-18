package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"test1/graph/model"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type getAgeResp struct {
	Count int64  `json:"count"`
	Name  string `json:"name"`
	Age   int    `json:"age"`
}

type getGenderResp struct {
	Count       int64   `json:"count"`
	Name        string  `json:"name"`
	Gender      string  `json:"gender"`
	Probability float32 `json:"probability"`
}

type getCountriesResp struct {
	Count   int              `json:"count"`
	Name    string           `json:"name"`
	Country []*model.Country `json:"country"`
}

type Service struct {
	Pool *pgxpool.Pool
	Rdb  *redis.Client
}

func New() *Service {
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	return &Service{
		Pool: pool,
		Rdb:  rdb,
	}
}

func (s *Service) AddUser(ctx context.Context, u model.NewUser) (*model.User, error) {
	var id int

	q := "insert into users(name, surname, patronymic, age, gender) values($1, $2, $3, $4, $5) returning id"
	row := s.Pool.QueryRow(ctx, q, u.Name, u.Surname, u.Patronymic, u.Age, u.Gender)
	err := row.Scan(&id)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	user := model.User{
		ID:         id,
		Name:       u.Name,
		Surname:    u.Surname,
		Patronymic: u.Patronymic,
		Age:        u.Age,
		Gender:     u.Gender,
		Country:    []*model.Country{},
	}

	return &user, nil
}

func (s *Service) GetUserById(ctx context.Context, id int) (*model.User, error) {
	res, err := s.Rdb.Get(ctx, fmt.Sprint(id)).Result()

	var u model.User

	if err != nil {
		q := "select id, name, surname, patronymic, age, gender from users where id = $1"
		row := s.Pool.QueryRow(ctx, q, id)
		row.Scan(&u.ID, &u.Name, &u.Surname, &u.Patronymic, u.Age, u.Gender)

		userBytes, err := json.Marshal(u)
		if err != nil {
			return nil, err
		}

		s.Rdb.Set(ctx, fmt.Sprint(id), userBytes, 0)

		return &u, nil
	}

	err = json.Unmarshal([]byte(res), &u)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (s *Service) ChangeUserById(ctx context.Context, u model.ChangeUser) error {
	q := "update usersset name = $1, surname = $2, patronymic = $3, age = $4, gender = $5) where id = $6"
	_, err := s.Pool.Exec(ctx, q, u.Name, u.Surname, u.Patronymic, u.Age, u.Gender, u.ID)

	return err
}

func (s *Service) DeleteUserById(ctx context.Context, id int) (*model.User, error) {
	var u model.User

	q := "delete from users where id = $1 returnint id, name, surname, patronymic, age, gender"
	row := s.Pool.QueryRow(ctx, q, id)
	err := row.Scan(&u.ID, &u.Name, &u.Surname, &u.Patronymic, u.Age, u.Gender)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (s *Service) GetUsers(ctx context.Context) ([]*model.User, error) {
	var users []*model.User

	q := "select id, name, surname, patronymic, age, gender from users"
	rows, err := s.Pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var u model.User

		err := rows.Scan(&u.ID, &u.Name, &u.Surname, &u.Patronymic, &u.Age, &u.Gender)
		if err != nil {
			return nil, err
		}

		users = append(users, &u)
	}

	return users, nil
}

func (s *Service) ProcessMessage(newUser *model.NewUser) (*model.User, error) {
	user := model.User{
		ID:         0,
		Name:       newUser.Name,
		Surname:    newUser.Surname,
		Patronymic: newUser.Patronymic,
		Age:        0,
		Gender:     "",
		Country:    []*model.Country{},
	}

	ageChan := make(chan int)
	genderChan := make(chan string)
	countriesChan := make(chan []*model.Country)

	go s.getAge(user.Name, ageChan)
	go s.getGender(user.Name, genderChan)
	go s.getCountries(user.Name, countriesChan)

	user.Age = <-ageChan
	user.Gender = <-genderChan
	user.Country = <-countriesChan

	return &user, nil
}

func (s *Service) getAge(name string, ch chan<- int) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", "https://api.agify.io", nil)
	if err != nil {
		return
	}
	q := req.URL.Query()
	q.Add("name", name)
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	var jsonResp getAgeResp
	json.Unmarshal(body, &jsonResp)
	log.Println(jsonResp)
	ch <- jsonResp.Age
}

func (s *Service) getGender(name string, ch chan<- string) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", "https://api.genderize.io", nil)
	if err != nil {
		return
	}
	q := req.URL.Query()
	q.Add("name", name)
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	var jsonResp getGenderResp
	json.Unmarshal(body, &jsonResp)
	log.Println(jsonResp)
	ch <- jsonResp.Gender
}

func (s *Service) getCountries(name string, ch chan<- []*model.Country) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", "https://api.nationalize.io", nil)
	if err != nil {
		return
	}
	q := req.URL.Query()
	q.Add("name", name)
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	var jsonResp getCountriesResp
	json.Unmarshal(body, &jsonResp)
	log.Println(jsonResp)
	ch <- jsonResp.Country
}
