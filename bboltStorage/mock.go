package bboltStorage

import (
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"testing"

	"github.com/runar-rkmedia/gabyoall/logger"
	"github.com/runar-rkmedia/skiver/types"
)

type mockDB struct {
	types.Storage
	L logger.AppLogger
}

func NewMockDB(t *testing.T) mockDB {
	l := logger.GetLoggerWithLevel("test", "fatal")
	tmpFile, err := ioutil.TempFile(os.TempDir(), "mockdb-skiver-")
	if err != nil {
		t.Fatal("Cannot create temporary file", err)
	}
	t.Logf("Created temporary db-file %s", tmpFile.Name())
	t.Cleanup(func() {
		t.Logf("Cleaned up temporary db-file %s", tmpFile.Name())
		os.Remove(tmpFile.Name())
	})
	bb, err := NewBbolt(l, tmpFile.Name(), nil, BBoltOptions{IDGenerator: &mockIdGenerator{}})
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
