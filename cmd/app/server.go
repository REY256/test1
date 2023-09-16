package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"test1/graph"
	"test1/graph/model"
	helpers "test1/internal"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
}

func main() {
	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		log.Fatal()
	}
	kafkaPort := os.Getenv("KAFKA_PORT")
	if kafkaPort == "" {
		log.Fatal()
	}
	topic := os.Getenv("TOPIC")
	err_topic := os.Getenv("ERR_TOPIC")

	conn, err := kafka.DialLeader(context.Background(), "tcp", "localhost:"+kafkaPort, topic, 0)
	if err != nil {
		log.Fatal("failed to dial leader:", err)
	}
	err_conn, err := kafka.DialLeader(context.Background(), "tcp", "localhost:"+kafkaPort, err_topic, 0)
	if err != nil {
		log.Fatal("failed to dial leader:", err)
	}
	fmt.Println("1")

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	go func() {
		b := make([]byte, 1e6)

		for {
			n, err := conn.Read(b)
			if err != nil {
				break
			}

			u := model.NewUser{}
			if err := json.Unmarshal(b[:n], &u); err != nil {
				err_conn.WriteMessages(kafka.Message{Value: []byte(b[:n])})
			}

			// обогащение сообщения
			user, err := helpers.ProcessMessage(&u)
			if err != nil {
				break
			}

			ctx := context.Background()

			tx, err := pool.Begin(ctx)
			if err != nil {
				break
			}

			_, err = tx.Exec(ctx, "insert into test_table(name, surname, patronymic, age, gender) values($1, $2, $3, $4, $5)", user.Name, user.Surname, user.Patronymic, user.Age, user.Gender)
			if err != nil {
				break
			}

			if err := tx.Commit(ctx); err != nil {
				break
			}

			fmt.Println(user)
		}

		if err := conn.Close(); err != nil {
			log.Fatal("failed to close connection:", err)
		}
	}()

	resolver := &graph.Resolver{Pool: pool, Rdb: rdb}

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", serverPort)
	log.Fatal(http.ListenAndServe(":"+serverPort, nil))

	r := chi.NewRouter()

	r.Get("/user", resolver.GetUserByIdHandler)
	r.Post("/add_user", resolver.AddUserHandler)
	r.Post("/change_user", resolver.ChangeUserHandler)
	r.Delete("/delete_user", resolver.DeleteUserByIdHandler)

	http.ListenAndServe("127.0.0.1:8080", r)

	fmt.Println("server online")
}
