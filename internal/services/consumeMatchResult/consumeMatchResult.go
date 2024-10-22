package consumeMatchResult

import (
	"context"
	"encoding/json"
	"log"

	"github.com/lukemakhanu/magic_carpet/internal/domains/messaging/rabbit"
	"github.com/lukemakhanu/magic_carpet/internal/domains/mrs"
	"github.com/lukemakhanu/magic_carpet/internal/domains/mrs/mrsMysql"
)

type ConsumeMatchResultConfiguration func(os *ConsumeMatchResultService) error

// ConsumeMatchResultService is a implementation of the ConsumeMatchResultService
type ConsumeMatchResultService struct {
	consume  *rabbit.QueueConsume
	mrsMysql mrs.MrsRepository
}

type Data struct {
	RawData string
}

// NewConsumeMatchResultService :
func NewConsumeMatchResultService(cfgs ...ConsumeMatchResultConfiguration) (*ConsumeMatchResultService, error) {
	os := &ConsumeMatchResultService{}
	for _, cfg := range cfgs {
		err := cfg(os)
		if err != nil {
			return nil, err
		}
	}
	return os, nil
}

func WithRabbitConsumeMatchResult(connectionString, queueName, connName, consumerName string) ConsumeMatchResultConfiguration {
	return func(m *ConsumeMatchResultService) error {
		consumer := rabbit.NewQueueConsume(connectionString, queueName, connName, consumerName)
		m.consume = consumer
		return nil
	}
}

// WithMysqlMatchesRepository : instantiates mysql to connect to matches interface
func WithMysqlMrsRepository(connectionString string) ConsumeMatchResultConfiguration {
	return func(os *ConsumeMatchResultService) error {
		m, err := mrsMysql.New(connectionString)
		if err != nil {
			return err
		}
		os.mrsMysql = m
		return nil
	}
}

// ConsumeMatchResult :
func (s *ConsumeMatchResultService) ConsumeMatchResult(ctx context.Context, ch chan Data) error {

	s.consume.Consume(func(data string) {
		// Push data into channel
		ch <- Data{RawData: data}
	})

	return nil
}

// SaveMatch : saves data into db
func (s *ConsumeMatchResultService) SaveMatchResult(ctx context.Context, ch chan Data) error {

	for p := range ch {

		var l mrs.TotalGoalCount
		if err := json.Unmarshal([]byte(p.RawData), &l); err != nil {
			log.Printf("Unable to total count: %v", err)
		} else {

			log.Printf("RoundNumberID : %d | CompetitionID : %s | StartTime : %s TotalGoals : %s | GoalCount : %s | RawScores : %s",
				l.RoundNumberID, l.CompetitionID, l.StartTime, l.TotalGoals, l.GoalCount, l.RawScores)

			dd, err := mrs.NewMrs(l.RoundNumberID, l.TotalGoals, l.GoalCount, l.CompetitionID, l.StartTime, l.RawScores)
			if err != nil {
				log.Printf("Err : %v", err)
			}

			lastID, err := s.mrsMysql.Save(ctx, *dd)
			if err != nil {
				log.Printf("Err : %v", err)
			}

			log.Printf("Last ID saved %d", lastID)

		}
	}

	return nil
}
