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
func (d DirectoryContext) CopyContext(context string) (string, error) {
	// Create GCS storage bucket
	bucket, bucketName, err := storage.CreateStorageBucket()
	if err != nil {
		return "", err
	}
	if err := storage.UploadContextToBucket(context, bucket); err != nil {
		return "", err
	}
	return bucketName, err
}

func (d DirectoryContext) GetFileFromContext(filename string) ([]byte, error) {

	return nil, nil
}

func (d DirectoryContext) CleanupContext() error {
	return nil
}
