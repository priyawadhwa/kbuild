package contexts

import (
	"github.com/priyawadhwa/kbuild/pkg/storage"
)

type DirectoryContext struct {
}

func (d DirectoryContext) Name() string {
	return "directory"
}

func (d DirectoryContext) CopyContext(context string) (string, error) {
	// Create GCS storage bucket
	if err := storage.CreateStorageBucket(context); err != nil {
		return err
	}

	return nil
}

func (d DirectoryContext) GetFileFromContext(filename string) ([]byte, error) {

	return nil, nil
}

func (d DirectoryContext) CleanupContext() error {
	return nil
}
