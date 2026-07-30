package cache

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var ErrCacheLevelNone error = errors.New("no cache for this key")

const (
	program_name       = "berichtsheft"
	metadata_file_name = "meta.json"
)

type FileCache struct {
	mu sync.RWMutex

	meta       map[string]Metadata
	baseFolder string
	ttl        time.Duration
	log        *slog.Logger
}

func NewFileCache(cacheID string, cacheTTL time.Duration) (*FileCache, error) {

	l := slog.Default().With("layer", "file-cache")

	userCache, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}

	var metadata map[string]Metadata
	baseFolder := filepath.Join(userCache, program_name, hashStr(cacheID))
	metadataPath := filepath.Join(baseFolder, metadata_file_name)

	if err := CreateDir(baseFolder); err != nil {
		return nil, fmt.Errorf("failed to create .cache directory. %w", err)
	}

	// Initialize the metadata
	if FileExists(metadataPath) {

		l.Debug("metadata file exists", "cache-id", cacheID)

		metafile, err := os.Open(metadataPath)
		if err != nil {
			return nil, fmt.Errorf("failed to open metadata file. %w", err)
		}
		defer CloseWithErrorLog(metafile)

		decoder := json.NewDecoder(metafile)
		err = decoder.Decode(&metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to read metadata. %w", err)
		}
	} else {

		l.Debug("metadata file does not exist", "cache-id", cacheID)

		// Clear the directory
		err = RemoveContents(baseFolder)
		if err != nil {
			return nil, fmt.Errorf("failed to remove contents of directory. %w", err)
		}

		metadata = make(map[string]Metadata)
		f, err := os.Create(metadataPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create metadata file. %w", err)
		}
		defer CloseWithErrorLog(f)
	}

	return &FileCache{
		meta:       metadata,
		baseFolder: baseFolder,
		log:        l,
		ttl:        cacheTTL,
	}, nil
}

func hashStr(key string) string {
	encrypted := sha256.Sum256([]byte(key))
	return hex.EncodeToString(encrypted[:])
}

func (f *FileCache) Close() error {
	return nil
}

func (f *FileCache) Exists(key string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if meta, exists := f.meta[key]; exists && meta.FileName != "" {
		f.log.Debug("cache does exist", "key", key, "fileName", meta.FileName)
		return true
	}
	f.log.Debug("cache does not exist", "key", key)
	return false
}

func (f *FileCache) getFilePath(key string) (string, error) {

	if data, exists := f.meta[key]; exists {
		return data.FileName, nil
	}

	return filepath.Join(f.baseFolder, hashStr(key)+".cache"), nil
}

func (f *FileCache) getMetadataFilePath() string {

	return filepath.Join(f.baseFolder, metadata_file_name)
}

func (f *FileCache) getFileForWrite(key string) (*os.File, error) {
	path, err := f.getFilePath(key)
	if err != nil {
		return nil, err
	}
	return os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
}

func (f *FileCache) getFileForRead(key string) (*os.File, error) {
	path, err := f.getFilePath(key)
	if err != nil {
		return nil, err
	}
	return os.Open(path)
}

func (f *FileCache) writeMetadataToFile() error {
	file, err := os.OpenFile(f.getMetadataFilePath(), os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open metadata file. %w", err)
	}
	defer CloseWithErrorLog(file)

	data, err := json.Marshal(f.meta)
	if err != nil {
		return fmt.Errorf("failed to marshal data. %w", err)
	}

	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write data into metadata file. %w", err)
	}

	return nil
}

func (f *FileCache) Set(key string, value []byte) error {
	file, err := f.getFileForWrite(key)
	if err != nil {
		return err
	}
	defer CloseWithErrorLog(file)

	_, err = file.Write(value)
	if err != nil {
		return err
	}

	f.mu.Lock()
	defer f.mu.Unlock()
	f.meta[key] = Metadata{RevalidationTime: time.Now(), FileName: file.Name()}

	return f.writeMetadataToFile()
}

func (f *FileCache) SetReader(key string, value io.Reader) (io.Reader, error) {
	file, err := f.getFileForWrite(key)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer CloseWithErrorLog(file)

	data, err := io.ReadAll(value)
	if err != nil {
		return nil, fmt.Errorf("failed to read from reader: %w", err)
	}

	err = f.Set(key, data)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(data), nil
}

func (f *FileCache) SetMetadata(key string, meta Metadata) error {
	f.meta[key] = meta
	return nil
}

func (f *FileCache) Get(key string) ([]byte, error) {
	file, err := f.getFileForRead(key)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	defer CloseWithErrorLog(file)

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read from reader: %w", err)
	}

	return data, nil
}

func (f *FileCache) GetReader(key string) (io.Reader, error) {
	file, err := f.getFileForRead(key)
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}
	defer CloseWithErrorLog(file)

	// Error
	b, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read from reader: %w", err)
	}

	return bytes.NewReader(b), nil
}

func (f *FileCache) GetMetadata(key string) (Metadata, error) {
	if data, exists := f.meta[key]; exists {
		return data, nil
	}

	return Metadata{}, fmt.Errorf("no entry for %s", key)
}

func (f *FileCache) IsValid(key string) (bool, error) {

	exists := f.Exists(key)
	if !exists {
		return false, nil
	}

	expires := f.meta[key].RevalidationTime.Add(f.ttl)

	if time.Now().After(expires) {
		f.log.Debug("invalid cache, beyond ttl", "key", key)
		return false, nil
	}

	f.log.Debug("valid cache", "key", key)
	return true, nil
}

func (f *FileCache) ExtendTTL(key string, duration time.Duration) error {
	return nil
}

func (f *FileCache) GetCacheFolder() string {
	return f.baseFolder
}
