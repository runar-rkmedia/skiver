package backup

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/bep/debounce"
	"github.com/dustin/go-humanize"
	"github.com/runar-rkmedia/go-common/logger"
	"github.com/runar-rkmedia/skiver/config"
	"github.com/runar-rkmedia/skiver/uploader"
)

type backuper struct {
	l logger.AppLogger
	// TODO: replace these structs with actual value
	config        map[string]config.BackupConfig
	uploaders     map[string]uploader.FileUploader
	stats         map[string]*BackupStats
	debouncedAuto func(f func())
	sync.Mutex
}

type BackupStats struct {
	LastUpdate time.Time
	Size       int64
	Hash       string
}

func NewBackHandler(l logger.AppLogger, cfg map[string]config.BackupConfig) *backuper {
	bak := backuper{
		l:         l,
		config:    cfg,
		stats:     map[string]*BackupStats{},
		uploaders: map[string]uploader.FileUploader{},
	}
	if len(cfg) == 0 {
		l.Fatal().Msg("no configs received")
	}

	shortestDuration := time.Hour * 1000
	for key, bkcfg := range cfg {
		if bkcfg.MaxInterval.Duration() < shortestDuration {
			shortestDuration = bkcfg.MaxInterval.Duration()
		}
		if bkcfg.FileName == "" {
			bkcfg.FileName = "skiver.bbolt"
			cfg[key] = bkcfg
		}
		switch {
		case bkcfg.S3 != nil:
			upl := uploader.NewS3Uplaoder(l, key, uploader.S3UploaderOptions{
				// TODO: use signed urls or ensure private buckets
				Endpoint:       bkcfg.S3.Endpoint,
				Region:         bkcfg.S3.Region,
				Bucket:         bkcfg.S3.BucketID,
				ProviderName:   bkcfg.S3.ProviderName,
				AccessKey:      bkcfg.S3.AccessKey,
				ForcePathStyle: bkcfg.S3.ForcePathStyle,
			}, bkcfg.S3.PrivateKey)
			bak.uploaders[key] = upl
			bak.stats[key] = nil
		default:
			l.Fatal().
				Str("key", key).
				Msg("error settup up uploader for endpoint; no valid configuration could be resolved for key. Currently, the only target available is S3, but that configuration was not provided")
		}
	}
	minDuration := time.Second
	if shortestDuration < minDuration {
		l.Warn().
			Str("shortestDuration", shortestDuration.String()).
			Str("minimumDuration", minDuration.String()).
			Msg("the shortest MaxInterval was larger than the minimum, and has been ignored")
		shortestDuration = minDuration
	}
	bak.debouncedAuto = debounce.New(shortestDuration)
	l.Debug().
		Str("debounceDuration", shortestDuration.String()).
		Msg("Database will be backed up after writes, at a debounced duration")

	return &bak
}

// Returns a list of keys of targets of which the backup eiter does not exist,
// or the backup is older than conf MaxInterval for that target
func (bak *backuper) BackupIsOlder(lastmodified time.Time) []string {
	olderKeys := []string{}
	bak.Lock()
	defer bak.Unlock()
	for k, v := range bak.stats {
		if v == nil {
			head, err := bak.uploaders[k].HeadFile(bak.config[k].FileName)
			if err != nil {
				bak.l.Error().
					Err(err).
					Str("key", k).
					Msg("Failed to check metadata on external backup")
				olderKeys = append(olderKeys, k)
				continue
			}
			if head.LastModified == nil {
				bak.l.Error().
					Str("key", k).
					Interface("metadata", head).
					Msg("Recieved metadata for external backup, but no LastModified")
				olderKeys = append(olderKeys, k)
				continue
			}
			v = &BackupStats{
				LastUpdate: *head.LastModified,
				Size:       *head.ContentLength,
			}
			bak.stats[k] = v
		}
		notBefore := v.LastUpdate.Add(time.Duration(bak.config[k].MaxInterval.Duration()))
		if lastmodified.After(notBefore) {
			olderKeys = append(olderKeys, k)
		}
	}
	return olderKeys
}

func (bak *backuper) BackupIsNotEqual(hash string) bool {
	bak.Lock()
	defer bak.Unlock()
	if len(bak.stats) == 0 {
		return true
	}
	for _, v := range bak.stats {
		if v == nil {
			return true
		}
		if v.Hash != hash {
			return true
		}
	}
	return false
}

// I find this kind of naming hilarious, please excuse me
type BackerUpper interface {
	Backup(io.Writer) (int64, error)
}

func Compress(r io.Reader, w io.Writer) (int, error) {
	gw := gzip.NewWriter(w)
	defer gw.Close()
	// bw := bufio.NewWriter(gw)
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return 0, err
	}
	return gw.Write(b)
	// return bw.ReadFrom(r)
}
func Decompress(r io.Reader, w io.Writer) (int, error) {
	gr, err := gzip.NewReader(r)
	if err != nil {
		return 0, err
	}
	defer gr.Close()
	b, err := ioutil.ReadAll(gr)
	if err != nil {
		return 0, err
	}
	return w.Write(b)
}
func (bak *backuper) AutoCreateBackupAndSaveIfRequired(source BackerUpper) error {
	bak.debouncedAuto(func() {
		bak.autoCreateBackupAndSaveIfRequired(source)
	})

	return nil

}

// Will check if backup is required depending on last time of backup, hash, etc,
// and then backup to all backup-sources if needed
func (bak *backuper) autoCreateBackupAndSaveIfRequired(source BackerUpper) error {
	l := bak.l
	targetKeys := bak.BackupIsOlder(time.Now())
	if len(targetKeys) == 0 {
		l.Debug().Msg("No targets require backup at this moment")
		return nil
	}

	l.Debug().
		Interface("targets", targetKeys).
		Int("targetCount", len(targetKeys)).
		Msg("Will now start backup to these targets")

	r := bytes.Buffer{}
	// www := gzip.NewWriter(&r)
	originalSize, err := source.Backup(&r)
	if err != nil {
		return err
	}

	// b, err := ioutil.ReadAll(&r)
	// if err != nil {
	// 	return err
	// }
	c := bytes.Buffer{}
	n, err := Compress(&r, &c)
	if err != nil {
		return err
	}
	readSeeker := bytes.NewReader(c.Bytes())

	bak.BackupIsNotEqual("foo")
	hash := sha256.New()
	hash.Write(r.Bytes())
	hashStr := hex.EncodeToString(hash.Sum(nil))
	// www.Close()
	if err != nil {
		l.Error().
			Str("hash", hashStr).
			Str("original", humanize.Bytes(uint64(originalSize))).
			Str("compressed", humanize.Bytes(uint64(n))).
			Err(err).Msg("error dumping backup to database")
		return err
	}
	if originalSize == 0 {
		err = fmt.Errorf("The backup created was empty (size was zero)")
		l.Error().
			Msg(err.Error())
		return err
	}
	l.Debug().
		Str("hash", hashStr).
		Str("original", humanize.Bytes(uint64(originalSize))).
		Str("compressed", humanize.Bytes(uint64(c.Len()))).
		Err(err).Msg("Initiating SaveBackup")
	written, err := bak.SaveBackup(targetKeys, time.Now(), int64(readSeeker.Len()), hashStr, readSeeker)
	if err != nil {
		l.Error().
			Str("written", humanize.Bytes(uint64(written))).
			Err(err).Msg("error writing backup")
		return err
	}
	l.Info().
		Str("hash", hashStr).
		Str("original", humanize.Bytes(uint64(originalSize))).
		Str("compressed", humanize.Bytes(uint64(written))).
		Err(err).Msg("database-backup saved")
	if written == 0 {
		err = fmt.Errorf("No data written")
		l.Error().
			Msg(err.Error())

	}
	l.Info().
		Str("written", humanize.Bytes(uint64(written))).
		Err(err).Msg("wrote backup")
	return nil
}

type HeadMeta struct {
	Hash         string
	Size         int64
	LastModified time.Time
}

func NewHeadMeta(hash string, size int64, lastmodified time.Time) HeadMeta {
	return HeadMeta{hash, size, lastmodified}
}
func NewHeadMetaFromMap(metadata map[string]*string) (HeadMeta, error) {
	h := HeadMeta{}
	if len(metadata) == 0 {
		return h, fmt.Errorf("Expected metadata to be set, but it was null")
	}
	hash, ok := metadata["Hash"]
	if !ok {
		return h, fmt.Errorf("Expected to find hash in metadata, but it was not found")
	}
	h.Hash = *hash
	size, err := strconv.ParseInt(*metadata["Size"], 10, 64)
	if err != nil {
		return h, fmt.Errorf("failed to parse size from metadata: %w", err)
	}
	h.Size = size
	lastmodified, err := time.Parse(time.RFC3339, *metadata["Lastmodified"])
	if err != nil {
		return h, fmt.Errorf("failed to parse lastmodified from metadata: %w", err)
	}
	h.LastModified = lastmodified
	return h, err
}
func (hm HeadMeta) ToMap() map[string]*string {
	return map[string]*string{
		"hash":         &hm.Hash,
		"size":         aws.String(strconv.FormatInt(hm.Size, 10)),
		"lastmodified": aws.String(hm.LastModified.UTC().Format(time.RFC3339)),
	}
}
func fileExists(filePath string) bool {
	_, error := os.Stat(filePath)
	return !errors.Is(error, os.ErrNotExist)
}
func (bak *backuper) WriteNewestBackup(filePath string) error {
	reader, err := bak.GetNewestBackup()
	if err != nil {
		return err
	}
	if reader == nil {
		return nil
	}
	l := bak.l
	tmp, err := os.CreateTemp(path.Dir(filePath), "skiver-bk")
	if err != nil {
		return fmt.Errorf("Failed to create temporary file for databasebackup: %w", err)
	}
	tmpPath := path.Join(tmp.Name())
	defer func() {
		tmp.Close()
		if fileExists(tmpPath) {
			os.Remove(tmpPath)
		}
	}()
	written, err := Decompress(reader, tmp)
	if err != nil {
		return fmt.Errorf("Failed to write to temporary file for databasebackup: %w (wrote %d)", err, written)
	}
	l.Info().
		Int("written", written).
		Str("tmpPath", tmpPath).
		Msg("Successfully retrieved database from backup")
	err = os.Rename(tmpPath, filePath)
	if err != nil {
		return fmt.Errorf("Failed to move temporary file (%s) to db-location (%s): %w", tmpPath, filePath, err)
	}
	return nil
}
func (bak *backuper) GetNewestBackup() (io.ReadCloser, error) {
	var latest *HeadMeta
	var latestKey string
	for k, upl := range bak.uploaders {
		if !bak.config[k].FetchOnStartup {
			bak.l.Debug().Str("targetKey", k).Msg("Ignoring backup for target, since FetchOnStartup is false for this target")
			continue
		}
		head, err := upl.HeadFile(bak.config[k].FileName)
		if err != nil {
			bak.l.Error().Str("targetKey", k).Err(err).Msg("Failed during HEAD-operation for backup")
			continue
		}
		hm, err := NewHeadMetaFromMap(head.Metadata)
		if err != nil {
			bak.l.Error().Str("targetKey", k).Err(err).Msg("Failed during metadata-extraction from HEAD-operation for backup")
			continue
		}
		if latest == nil {
			latest = &hm
			latestKey = k
			continue
		}
		if latest.LastModified.Before(hm.LastModified) {
			latest = &hm
			latestKey = k
		}
	}
	if latest == nil {
		return nil, nil
	}
	g, err := bak.uploaders[latestKey].GetFile(bak.config[latestKey].FileName)
	if err != nil {
		bak.l.Error().Err(err).
			Str("targetKey", latestKey).
			Msg("faild to get latest backup")
		return nil, fmt.Errorf("failed to get latest backup: %w", err)
	}
	return g.Body, err

}
func (bak *backuper) SaveBackup(targetKeys []string, lastmodified time.Time, size int64, hash string, r io.ReadSeeker) (int64, error) {
	if size <= 0 {
		return 0, fmt.Errorf("cannot backup with a zero-length-file")
	}
	b := bytes.Buffer{}
	meta := NewHeadMeta(hash, size, lastmodified)
	metaMap := meta.ToMap()
	fileOptions := uploader.AddFileOptions{
		Metadata: metaMap,
	}
	for _, key := range targetKeys {
		upl := bak.uploaders[key]
		uploadType, err := upl.AddPublicFile(bak.config[key].FileName, r, size, "application/gzip", "", fileOptions)
		if err != nil {
			bak.l.Error().
				Err(err).
				Interface("uploadType", uploadType).
				Str("hash", hash).
				Str("key", key).
				Str("size", humanize.Bytes(uint64(size))).
				Str("Identifier", upl.Identifier()).
				Msg("Failed during upload")

			return 0, err
		}
		bak.l.Debug().
			Interface("uploadType", uploadType).
			Str("hash", hash).
			Str("key", key).
			Str("size", humanize.Bytes(uint64(size))).
			Str("Identifier", upl.Identifier()).
			Msg("Upload successful")

		r.Seek(0, io.SeekStart)

	}
	n, err := b.ReadFrom(r)
	os.WriteFile("./backup-ignoreme", b.Bytes(), 0644)
	return n, err
	// TODO: use cancellation, in case a new write to the backup is already happening, and we want to use that instead.
	// TODO: create multiwriter for each s3-endpoint
	// OR: write to a buffer, and use that buffer to write to each s3-endpoint sequentially.
}
