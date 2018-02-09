package dest

import (
	"github.com/priyawadhwa/kbuild/pkg/storage"
)

type BucketContext struct {
	bucketName string
}

// GetFilesFromSource gets the files at path from the GCS storage bucket
func (b BucketContext) GetFilesFromSource(path string) (map[string][]byte, error) {
	return storage.GetFilesFromStorageBucket(b.bucketName, path)
}
func (b BucketContext) CleanupContext() error {
	return storage.DeleteStorageBucket(b.bucketName)
}
