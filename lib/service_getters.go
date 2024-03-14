package lib

func (s *Service) GetQueueNameByRef(queueRef string) *string {
	queueName, found := s.Config.Queues[queueRef]
	if found == false {
		return nil
	}

	return &queueName
}
