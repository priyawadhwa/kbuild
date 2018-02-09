package contexts

type Context interface {
	Name() string
	CopyFilesToContext(files []string) (string, error)
	GetFilesFromSource(path, source string) (map[string][]byte, error)
	CleanupContext() error
}

var Contexts = map[string]Context{
	"directory": DirectoryContext{},
}

func GetContext(context string) Context {
	return DirectoryContext{}
}
