package parser

//go:generate go run ../../../utils/typeGenerator stack ClassStack *Class

type ClassStack []*Class

func NewClassStack() *ClassStack {
	s := make(ClassStack, 0)
	return &s
}

func (s *ClassStack) Push(value *Class) {
	*s = append(*s, value)
}

func (s *ClassStack) Pop() (*Class, bool) {
	if s.IsEmpty() {
		return nil, false
	}
	elm := (*s)[len(*s)-1]
	*s = (*s)[:len(*s)-1]
	return elm, true
}

func (s *ClassStack) Peek() (*Class, bool) {
	if s.IsEmpty() {
		return nil, false
	}
	return (*s)[len(*s)-1], true
}

func (s *ClassStack) IsEmpty() bool {
	return len(*s) == 0
}
