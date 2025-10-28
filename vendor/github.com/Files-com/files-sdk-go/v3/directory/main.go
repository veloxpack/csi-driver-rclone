package directory

type Type string

var (
	Dir   = Type("directory")
	File  = Type("file")
	Files = Type("files")
)
