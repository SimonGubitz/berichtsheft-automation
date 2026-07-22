package cache_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"berichtsheft/internal/cache"
)

const (
	testKey           = "test_key"
	want              = "test string for cache"
	testKeyCustomFile = "test_key_with_custom_file"
)

func TestFileCache(t *testing.T) {

	cacheTestID := "cache_test_id"

	fileCache, err := cache.NewFileCache(cacheTestID)
	if err != nil {
		t.Fatalf("failed to create file cache. %v", err)
	}

	err = fileCache.Set(testKey, []byte(want))
	if err != nil {
		t.Fatalf("failed to set file cache. %v", err)
	}

	have, err := fileCache.Get(testKey)
	if err != nil {
		t.Fatalf("failed to get file cache. %v", err)
	}

	if want != string(have) {
		t.Fatal("cache returned false values")
	}

	haveReader, err := fileCache.GetReader(testKey)
	if err != nil {
		t.Fatalf("could not get reader from file cache. %v", err)
	}

	have, err = io.ReadAll(haveReader)
	if err != nil {
		t.Fatalf("could not read on returned file cache reader. %v", err)
	}
	if want != string(have) {
		t.Fatal("cache returned false values")
	}
}

func TestFileMetadata(t *testing.T) {

	testCacheID := "test_cache_id"

	tmpFile, err := os.CreateTemp("", "*")
	if err != nil {
		t.Fatalf(". %v", err)
	}
	defer func() {
		if err := tmpFile.Close(); err != nil {
			t.Fatal(err)
		}

		if err := os.Remove(tmpFile.Name()); err != nil {
			t.Fatal(err)
		}
	}()

	_, err = tmpFile.Write([]byte(want))
	if err != nil {
		t.Fatalf("failed to prepare mock contents. %v", err)
	}

	wantMeta := cache.Metadata{
		RevalidationTime: time.Now(),
		FileName:         tmpFile.Name(),
	}

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(cacheDir, "berichtsheft", hashStr(testCacheID), "meta.json")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		t.Fatalf("failed to open metadata file. %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Fatalf("failed to close metadata file. %v", err)
		}
	}()

	// write the tmp into the meta.json, explicitly bypassing the SetMetadata API
	tmp := make(map[string]cache.Metadata)
	tmp[testKeyCustomFile] = wantMeta
	encoder := json.NewEncoder(f)
	err = encoder.Encode(tmp)
	if err != nil {
		t.Fatalf("failed to encode into metadata file. %v", err)
	}

	fileCache, err := cache.NewFileCache(testCacheID)
	if err != nil {
		t.Fatalf("failed to create file cache. %v", err)
	}

	haveMeta, err := fileCache.GetMetadata(testKeyCustomFile)
	if err != nil {
		t.Fatalf("could not get metadata from cache. %v", err)
	}

	t.Logf("Metadata: %+v", haveMeta)
	if !haveMeta.RevalidationTime.Equal(wantMeta.RevalidationTime) {
		t.Fatalf("revalidation times differ. want=%v have=%v", wantMeta.RevalidationTime, haveMeta.RevalidationTime)
	}

	if haveMeta.FileName != wantMeta.FileName {
		t.Fatalf("file names differ. want=%q have=%q", wantMeta.FileName, haveMeta.FileName)
	}

	have, err := fileCache.Get(testKeyCustomFile)
	if err != nil {
		absPath, errAbs := filepath.Abs(fileCache.GetCacheFolder())
		if errAbs != nil {
			t.Fatalf("failed to get absolute path on cache folder. err: %+v", errAbs)
		}

		t.Fatalf("could not get file for custom cache. key: %s, absolute path: %s. err: %+v", testKeyCustomFile, absPath, err)
	}

	haveStr := strings.TrimSuffix(string(have), "\n")
	if want != haveStr {
		t.Fatalf("cache returned false values. Want: \"%s\", Have: \"%s\"", want, haveStr)
	}
}

// func TestFileCacheUniqueID(t *testing.T) {
//
// 	cacheID := "duplicate_cache_id"
//
// 	_, err := cache.NewFileCache(cacheID)
// 	if err != nil {
// 		t.Fatalf("failed to create valid first cache. %v", err)
// 	}
//
// 	_, err = cache.NewFileCache(cacheID)
// 	if err == nil {
// 		t.Fatal("failed to throw an error on duplicate file id")
// 	}
// }

func hashStr(key string) string {
	encrypted := sha256.Sum256([]byte(key))
	return hex.EncodeToString(encrypted[:])
}
