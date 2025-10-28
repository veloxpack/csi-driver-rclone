package direction

type Direction struct {
	name   string
	Symbol string
}

var (
	DownloadType = Direction{name: "download", Symbol: "⬆"}
	UploadType   = Direction{name: "upload", Symbol: "⬆"}
)

func (t *Direction) Name() string {
	return t.name
}
