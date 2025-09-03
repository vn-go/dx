package expr

type stack[T any] struct {
	data []T
}

func (s *stack[T]) Push(v T) {
	s.data = append(s.data, v)
}

func (s *stack[T]) Pop() (T, bool) {
	if len(s.data) == 0 {
		var zero T
		return zero, false
	}
	val := s.data[len(s.data)-1]
	s.data = s.data[:len(s.data)-1]
	return val, true
}

func (s *stack[T]) Peek() (T, bool) {
	if len(s.data) == 0 {
		var zero T
		return zero, false
	}
	return s.data[len(s.data)-1], true
}

func (s *stack[T]) IsEmpty() bool {
	return len(s.data) == 0
}
