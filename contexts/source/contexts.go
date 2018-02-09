package source

type Context interface {
	Name() string
	CopyFilesToContext(files []string) (string, error)
}

func GetContext(context string) Context {
	return DirectoryContext{}
}
