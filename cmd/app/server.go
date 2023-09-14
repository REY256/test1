package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"test1/graph"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/joho/godotenv"
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
	topic := os.Getenv("MAIN_TOPIC")

	conn, err := kafka.DialLeader(context.Background(), "tcp", "localhost:"+kafkaPort, topic, 0)
	if err != nil {
		log.Fatal("failed to dial leader:", err)
	}

	batch := conn.ReadBatch(10e3, 1e6) // fetch 10KB min, 1MB max
	b := make([]byte, 10e3)            // 10KB max per message
	go func() {
		for {
			n, err := batch.Read(b)
			if err != nil {
				break
			}
			fmt.Println(string(b[:n]))
		}

		if err := batch.Close(); err != nil {
			log.Fatal("failed to close batch:", err)
		}

		if err := conn.Close(); err != nil {
			log.Fatal("failed to close connection:", err)
		}
	}()

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{}}))

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", serverPort)
	log.Fatal(http.ListenAndServe(":"+serverPort, nil))
}
