package dest

type Context interface {
	GetFilesFromSource(path string) (map[string][]byte, error)
	CleanupContext() error
}

func GetContext(source string) Context {
	return BucketContext{bucketName: source}
}
