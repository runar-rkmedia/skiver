package bboltStorage

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/dustin/go-humanize"
	bolt "go.etcd.io/bbolt"
)

// Backups the database to a writer.
// Note that bbolt seems to always change the file (statistics?),
// so creating a hash of the the database does not
func (s *BBolter) Backup(w io.Writer) (int64, error) {
	var originalSize int64 = -1
	var compactSize int64 = -1
	s.DB.View(func(tx *bolt.Tx) error {
		originalSize = tx.Size()
		return nil
	})
	s.l.Info().Msg("Creating backup of database, by first compacting the current database to a new temporary database")
	compactPath := "__compact-backup.bbolt"
	if fileExists(compactPath) {
		if err := os.Remove(compactPath); err != nil {
			return 0, err
		}
	}
	defer os.Remove(compactPath)

	compactDb, err := s.copyCompact(compactPath)
	if compactDb != nil {
		defer compactDb.Close()
	}
	if err != nil {
		return 0, err
	}
	var written int64
	err = compactDb.View((func(tx *bolt.Tx) error {
		compactSize = tx.Size()
		n, err := tx.WriteTo(w)
		written = n
		return err
	}))
	if err != nil {
		return written, err
	}
	if written == 0 {
		err = fmt.Errorf("The data written was of zero size")
		return written, err
	}
	s.l.Debug().
		Str("originalSize", humanize.Bytes(uint64(originalSize))).
		Str("compactSize", humanize.Bytes(uint64(compactSize))).
		Str("percentage", fmt.Sprintf("%2f", float64(compactSize)/float64(originalSize)*100)).
		Msg("Compact-size-difference")
	return written, err
}

func fileExists(filePath string) bool {
	_, error := os.Stat(filePath)
	return !errors.Is(error, os.ErrNotExist)
}
