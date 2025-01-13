package structx

type set[T comparable] struct {
	keys   map[T]int
	values []any
}

// Set ...
func Set[T comparable]() *set[T] {
	return &set[T]{
		keys:   make(map[T]int),
		values: make([]any, 0),
	}
}

func (s *set[T]) Set(key T, value any) {
	if x, ok := s.keys[key]; ok {
		s.values[x] = value
		return
	}
	s.keys[key] = len(s.values)
	s.values = append(s.values, value)
}

func (s *set[T]) Get(key T) any {

	x, ok := s.keys[key]
	if !ok {
		return nil
	}
	return s.values[x]
}

func (s *set[T]) Del(key T, value any) {

}

func (s *set[T]) Exist(key T) bool {
	_, ok := s.keys[key]
	return ok
}

func (s *set[T]) Values() []any {
	return s.values
}

func (s *set[T]) SortBy(field string) []any {
	return s.values
}
