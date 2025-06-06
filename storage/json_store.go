package storage

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
)

type JSONStorage struct {
	path string
}

func NewJSONStorage(path string) Storage {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			panic(err)
		}
	}

	return &JSONStorage{
		path: path,
	}
}

func (s *JSONStorage) SaveSession(ctx context.Context, session *Session) error {
	// create file
	filePath := filepath.Join(s.path, session.ID+".json")
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(session)
}

func (s *JSONStorage) FetchSession(ctx context.Context, id string) (*Session, error) {
	filePath := filepath.Join(s.path, id+".json")
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var session Session
	if err := json.NewDecoder(file).Decode(&session); err != nil {
		return nil, err
	}

	return &session, nil
}

func (s *JSONStorage) UpdateSession(ctx context.Context, session *Session) error {
	filePath := filepath.Join(s.path, session.ID+".json")
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(session)
}

func (s *JSONStorage) DeleteSession(ctx context.Context, id string) error {
	filePath := filepath.Join(s.path, id+".json")
	if err := os.Remove(filePath); err != nil {
		return err
	}

	return nil
}
