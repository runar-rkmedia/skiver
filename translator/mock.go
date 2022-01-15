package translator

import "fmt"

type MockTranlator struct{}

func (m *MockTranlator) Translate(text, from, to string) (string, error) {
	return fmt.Sprintf("[Mock] This text is in '%s', translated from '%s': %s", to, from, text), nil
}
