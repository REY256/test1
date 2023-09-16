package helpers

import (
	"encoding/json"
	"io"
	"net/http"
	"test1/graph/model"
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

func ProcessMessage(newUser *model.NewUser) (*model.User, error) {
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

	go getAge(user.Name, ageChan)
	go getGender(user.Name, genderChan)
	go getCountries(user.Name, countriesChan)

	user.Age = <-ageChan
	user.Gender = <-genderChan
	user.Country = <-countriesChan

	return &user, nil
}

func getAge(name string, ch chan<- int) {
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
	if err := json.Unmarshal(body, &jsonResp); err != nil {
		return
	}
	ch <- jsonResp.Age
}

func getGender(name string, ch chan<- string) {
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
	if err := json.Unmarshal(body, &jsonResp); err != nil {
		return
	}
	ch <- jsonResp.Gender
}

func getCountries(name string, ch chan<- []*model.Country) {
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
	if err := json.Unmarshal(body, &jsonResp); err != nil {
		return
	}
	ch <- jsonResp.Country
}
