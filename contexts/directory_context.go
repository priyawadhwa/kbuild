package contexts

import (
	"github.com/priyawadhwa/kbuild/pkg/storage"
)

type DirectoryContext struct {
}

func (d DirectoryContext) Name() string {
	return "directory"
}

// Copy local directory into a GCS storage bucket
func (d DirectoryContext) CopyFilesToContext(files []string) (string, error) {
	// Create GCS storage bucket
	bucket, bucketName, err := storage.CreateStorageBucket()
	if err != nil {
		return "", err
	}
	if err := storage.UploadContextToBucket(files, bucket); err != nil {
		return "", err
	}
	return bucketName, err
}

func (d DirectoryContext) GetFilesFromSource(path, source string) (map[string][]byte, error) {
	return storage.GetFilesFromStorageBucket(source, path)
}

func (d DirectoryContext) CleanupContext() error {
	return nil
}
