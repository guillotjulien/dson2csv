package internal

import "errors"

type Stack []string

func (s Stack) Push(v string) Stack {
	return append(s, v)
}

func (s Stack) Pop() (Stack, error) {
	l := len(s)
	if l == 0 {
		return nil, errors.New("Stack is empty")
	}
	return s[:l-1], nil
}

func (s Stack) Peek() string {
	l := len(s)
	if l == 0 {
		return ""
	}
	return s[l-1]
}
