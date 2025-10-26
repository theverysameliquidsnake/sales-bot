package types

type Set struct {
	elements map[string]struct{}
}

func NewSet() *Set {
	return &Set{
		elements: make(map[string]struct{}),
	}
}

func (s *Set) Add(value string) {
	s.elements[value] = struct{}{}
}

func (s *Set) Values() []string {
	var keys []string
	for key := range s.elements {
		keys = append(keys, key)
	}

	return keys
}
