package oauth

import (
	"os"
)

// OSFileReader implements FileReader using the OS filesystem
type OSFileReader struct{}

func (r *OSFileReader) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// OSFileWriter implements FileWriter using the OS filesystem
type OSFileWriter struct{}

func (w *OSFileWriter) WriteFile(path string, data []byte, perm uint32) error {
	return os.WriteFile(path, data, os.FileMode(perm))
}
