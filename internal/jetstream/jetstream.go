package jetstream

import (
	"log"

	utils "github.com/burakmert236/goodswipe/internal/utils"
	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
)

const STREAM_NAME string = "USER_ACTIVITY"

type NATS struct {
	NatsConnection   *nats.Conn
	JetStreamContext nats.JetStreamContext
}

type Subjects int

const (
	TournamentCompleted Subjects = iota
	UserLevelUp
)

var subjectNames = map[Subjects]string{
	TournamentCompleted: "tournament.completed",
	UserLevelUp:         "user_events.level_up",
}

func (s Subjects) String() string {
	return subjectNames[s]
}

func InitNats() *NATS {
	natsURL := utils.GetEnvRequired("NATS_URL")

	nc, err := nats.Connect(natsURL,
		nats.MaxReconnects(-1),
		nats.ReconnectWait(2e9),
	)
	if err != nil {
		log.Fatalf("[NATS] Failed to connect: %v", err)
	}

	js, err := nc.JetStream()
	if err != nil {
		log.Fatalf("[NATS] Failed to init JetStream: %v", err)
	}

	var subjectNamesArr = make([]string, 0, len(subjectNames))
	for _, v := range subjectNames {
		subjectNamesArr = append(subjectNamesArr, v)
	}

	js.AddStream(&nats.StreamConfig{
		Name:     STREAM_NAME,
		Subjects: subjectNamesArr,
	})

	log.Println("NATS JetStream initialized")
	return &NATS{NatsConnection: nc, JetStreamContext: js}
}

func (n *NATS) CloseNATS() {
	n.NatsConnection.Drain()
}

func (n *NATS) Publish(subjectName string, event proto.Message) error {
	data, err := proto.Marshal(event)
	if err != nil {
		return err
	}

	var publishError = n.NatsConnection.Publish(subjectName, data)
	if publishError != nil {
		return publishError
	}

	return nil
}
