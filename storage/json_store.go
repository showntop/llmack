package storage

import (
	"context"
	"encoding/json"
	"io"
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
	// 如果不存在就创建
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		_, err := os.Create(filePath)
		if err != nil {
			return nil, err
		}
		return &Session{
			ID: id,
		}, nil
	}
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	if len(content) == 0 {
		return &Session{
			ID: id,
		}, nil
	}

	var session Session
	if err := json.Unmarshal(content, &session); err != nil {
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
