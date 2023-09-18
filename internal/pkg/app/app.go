package app

import (
	"log"
	"net/http"
	"os"
	"test1/graph"
	"test1/internal/app/endpoint"
	"test1/internal/app/kafka"
	"test1/internal/app/service"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi/v5"
)

type App struct {
	rest  *endpoint.Endpoint
	gql   *graph.Resolver
	mux   *chi.Mux
	kafka *kafka.Kafka
}

func New() (*App, error) {
	service := service.New()

	a := &App{
		rest:  endpoint.New(service),
		gql:   graph.New(service),
		kafka: kafka.New(service),
		mux:   chi.NewRouter(),
	}

	a.mux.Get("/users", a.rest.GetUsersHandler)
	a.mux.Get("/user", a.rest.GetUserByIdHandler)
	a.mux.Post("/add_user", a.rest.AddUserHandler)
	a.mux.Post("/change_user", a.rest.ChangeUserHandler)
	a.mux.Delete("/delete_user", a.rest.DeleteUserByIdHandler)

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: a.gql}))

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	return a, nil
}

func (a *App) Run() {
	gqlServerPort := os.Getenv("GQL_SERVER_PORT")
	if gqlServerPort == "" {
		log.Fatal()
	}

	httpServerPort := os.Getenv("HTTP_SERVER_PORT")
	if httpServerPort == "" {
		log.Fatal()
	}

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", gqlServerPort)

	go func() {
		log.Fatal(http.ListenAndServe(":"+gqlServerPort, nil))
	}()

	a.kafka.Run()

	log.Fatal(http.ListenAndServe("localhost:"+httpServerPort, a.mux))
}
