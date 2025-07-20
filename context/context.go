package context

type Context struct {
	Type     string
	Content  string
	Metadata map[string]any
}
