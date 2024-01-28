package lib

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func (s *Service) GetAuthToken() string {
	return s.config.AuthToken
}

func (s *Service) GetValidationUrl() string {
	return s.config.JwtValidationUrl
}

func (s *Service) GetQueueNameByRef(queueRef string) *string {
	queueName, found := s.config.Queues[queueRef]
	if found == false {
		return nil
	}

	return &queueName
}

func (s *Service) StressTestAllowed() bool {
	return s.config.AllowStressTest
}

func (s *Service) ExecWithDb(req *Request, dbExec func(*pgxpool.Pool) error) error {

	dbpool, connErr := pgxpool.New(req.SentryContext, s.config.DbUrl)

	if connErr != nil {
		CaptureSentryException(fmt.Sprintf("%s Error fetching DB connection %s", req.ID, connErr.Error()))
		return connErr
	}

	if dbpool == nil {
		CaptureSentryException(fmt.Sprintf("%s Error creating db connection as dbpool returned nil", req.ID))
		return errors.New("DB Error")
	}

	defer dbpool.Close()

	return dbExec(dbpool)
}
