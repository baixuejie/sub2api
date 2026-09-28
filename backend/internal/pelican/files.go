package pelican

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

// Files are private, never mounted as a static web directory. Authenticated
// handlers return JSON and the browser renders only the sanitized sandbox copy.
type artifactFiles struct{ dir string }

func defaultArtifactFiles() *artifactFiles {
	dir := os.Getenv("PELICAN_ARTIFACT_DIR")
	if dir == "" {
		data := os.Getenv("DATA_DIR")
		if data == "" {
			data = "data"
		}
		dir = filepath.Join(data, "pelican")
	}
	return &artifactFiles{dir: dir}
}

var artifactPathPattern = regexp.MustCompile(`^[1-9][0-9]*/(source|preview)-[0-9a-f]{64}\.(txt|html)$`)

func (f *artifactFiles) open() (*os.Root, error) {
	if err := os.MkdirAll(f.dir, 0700); err != nil {
		return nil, err
	}
	return os.OpenRoot(f.dir)
}

// Fail before paying for a model call when the artifact volume is unwritable.
func (f *artifactFiles) checkWritable() error {
	root, err := f.open()
	if err != nil {
		return err
	}
	defer root.Close()
	var nonce [16]byte
	if _, err = rand.Read(nonce[:]); err != nil {
		return err
	}
	name := ".write-check-" + hex.EncodeToString(nonce[:])
	file, err := root.OpenFile(name, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer root.Remove(name)
	_, err = io.WriteString(file, "pelican")
	if err == nil {
		err = file.Sync()
	}
	closeErr := file.Close()
	if err == nil {
		err = closeErr
	}
	return err
}

func (f *artifactFiles) write(id int64, kind, content string) (string, error) {
	if id <= 0 || (kind != "source" && kind != "preview") || len(content) > MaxArtifactBytes {
		return "", errors.New("invalid artifact")
	}
	root, err := f.open()
	if err != nil {
		return "", err
	}
	defer root.Close()
	dir := strconv.FormatInt(id, 10)
	if err := root.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	ext := ".html"
	if kind == "source" {
		ext = ".txt"
	}
	path := dir + "/" + kind + "-" + contentHash(content) + ext
	var nonce [16]byte
	if _, err = rand.Read(nonce[:]); err != nil {
		return "", err
	}
	tmp := dir + "/.tmp-" + hex.EncodeToString(nonce[:])
	file, err := root.OpenFile(tmp, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return "", err
	}
	defer root.Remove(tmp)
	_, err = io.WriteString(file, content)
	if err == nil {
		err = file.Sync()
	}
	closeErr := file.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		return "", err
	}
	if err = root.Rename(tmp, path); err != nil {
		return "", err
	}
	// Verify before clearing legacy database content or committing its address.
	saved, err := f.read(path)
	if err != nil {
		return "", err
	}
	if saved != content {
		return "", errors.New("artifact verification failed")
	}
	d, err := root.Open(dir)
	if err != nil {
		return "", err
	}
	err = d.Sync()
	d.Close()
	if err != nil {
		return "", err
	}
	return path, nil
}

func (f *artifactFiles) read(path string) (string, error) {
	if !artifactPathPattern.MatchString(path) {
		return "", ErrNotFound
	}
	root, err := f.open()
	if err != nil {
		return "", err
	}
	defer root.Close()
	file, err := root.Open(path) // os.Root also prevents symlink traversal outside dir.
	if errors.Is(err, os.ErrNotExist) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() || info.Size() > MaxArtifactBytes {
		return "", fmt.Errorf("invalid artifact file")
	}
	data, err := io.ReadAll(io.LimitReader(file, MaxArtifactBytes+1))
	if err != nil {
		return "", err
	}
	if len(data) > MaxArtifactBytes {
		return "", errors.New("artifact too large")
	}
	expected := filepath.Base(path)
	if expected != "source-"+contentHash(string(data))+".txt" && expected != "preview-"+contentHash(string(data))+".html" {
		return "", errors.New("artifact checksum mismatch")
	}
	return string(data), nil
}

func (f *artifactFiles) removeRun(id int64) error {
	if id <= 0 {
		return errors.New("invalid artifact id")
	}
	root, err := f.open()
	if err != nil {
		return err
	}
	defer root.Close()
	return root.RemoveAll(strconv.FormatInt(id, 10))
}
