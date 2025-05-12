package storage

import "context"

type Storage interface {
	SaveSession(ctx context.Context, session *Session) error
	FetchSession(ctx context.Context, id string) (*Session, error)
	UpdateSession(ctx context.Context, session *Session) error
	DeleteSession(ctx context.Context, id string) error
}

type NoneStorage struct {
}

func (s *NoneStorage) SaveSession(ctx context.Context, session *Session) error {
	return nil
}

func (s *NoneStorage) FetchSession(ctx context.Context, id string) (*Session, error) {
	return nil, nil
}

func (s *NoneStorage) UpdateSession(ctx context.Context, session *Session) error {
	return nil
}

func (s *NoneStorage) DeleteSession(ctx context.Context, id string) error {
	return nil
}
