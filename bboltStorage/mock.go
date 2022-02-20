package bboltStorage

import (
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"testing"

	"github.com/runar-rkmedia/go-common/logger"
	"github.com/runar-rkmedia/skiver/types"
)

type mockDB struct {
	types.Storage
	L logger.AppLogger
}

func NewMockDB(t *testing.T) mockDB {
	t.Helper()
	l := logger.GetLoggerWithLevel("test", "fatal")
	tmpFile, err := ioutil.TempFile(os.TempDir(), "mockdb-skiver-")
	if err != nil {
		t.Fatal("Cannot create temporary file", err)
	}
	t.Cleanup(func() {
		os.Remove(tmpFile.Name())
	})
	bb, err := NewBbolt(l, tmpFile.Name(), nil, BBoltOptions{IDGenerator: &mockIdGenerator{}})
	if err != nil {
		t.Fatal(err)
	}
	return mockDB{&bb, l}
}

type mockIdGenerator struct {
	counter int
	sync.RWMutex
}

func (m *mockIdGenerator) CreateUniqueID() string {
	m.Lock()
	defer m.Unlock()
	m.counter++
	return fmt.Sprintf("id-%d", m.counter)
}

func (m *mockDB) StandardSeed() error {
	org, err := types.SeedUsers(m, nil, func(s string) ([]byte, error) { return []byte("mock-" + s), nil })
	if err != nil {
		return fmt.Errorf("Failed to seed users %w", err)
	}
	if org != nil {
		err = types.SeedLocales(m, org.ID, nil)
		if err != nil {
			return fmt.Errorf("Failed to seed Locale %w", err)
		}
	}

	return nil

}
