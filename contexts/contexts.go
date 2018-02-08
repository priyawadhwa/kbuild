package contexts

type Context interface {
	Name() string
	CopyContext(context string) (string, error)
	GetFileFromContext(filename string) ([]byte, error)
	CleanupContext() error
}

var Contexts = map[string]Context{
	"directory": DirectoryContext{},
}
