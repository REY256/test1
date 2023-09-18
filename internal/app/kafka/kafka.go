package kafka

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"test1/graph/model"
	"test1/internal/app/service"

	"github.com/segmentio/kafka-go"
)

type Kafka struct {
	s            *service.Service
	KafkaConn    *kafka.Conn
	KafkaErrConn *kafka.Conn
}

func New(s *service.Service) *Kafka {
	ctx := context.Background()

	kafkaPort := os.Getenv("KAFKA_PORT")
	if kafkaPort == "" {
		log.Fatal()
	}
	topic := os.Getenv("TOPIC")
	err_topic := os.Getenv("ERR_TOPIC")

	conn, err := kafka.DialLeader(ctx, "tcp", "localhost:"+kafkaPort, topic, 0)
	if err != nil {
		log.Fatal("failed to dial leader:", err)
	}
	errConn, err := kafka.DialLeader(ctx, "tcp", "localhost:"+kafkaPort, err_topic, 0)
	if err != nil {
		log.Fatal("failed to dial leader:", err)
	}

	return &Kafka{
		s:            s,
		KafkaConn:    conn,
		KafkaErrConn: errConn,
	}
}

func (k *Kafka) Run() {
	go func() {
		b := make([]byte, 1e6)

		for {
			n, err := k.KafkaConn.Read(b)
			if err != nil {
				log.Println(err)
				break
			}
			msg := b[:n]

			u := model.NewUser{}
			if err := json.Unmarshal(msg, &u); err != nil {
				_, err := k.KafkaErrConn.WriteMessages(kafka.Message{Value: msg})
				log.Println("err unmarshal user")
				if err != nil {
					log.Println(err)
					break
				}
			}

			// обогащение сообщения
			user, err := k.s.ProcessMessage(&u)
			if err != nil {
				log.Println(err)
				break
			}
			log.Println(user)

			ctx := context.Background()

			q := "insert into users(name, surname, patronymic, age, gender) values($1, $2, $3, $4, $5)"
			_, err = k.s.Pool.Exec(ctx, q, user.Name, user.Surname, user.Patronymic, user.Age, user.Gender)
			if err != nil {
				log.Println(err)
				break
			}
		}

		if err := k.KafkaConn.Close(); err != nil {
			log.Fatal("failed to close connection:", err)
		}
	}()
}
