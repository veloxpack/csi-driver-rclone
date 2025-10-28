package status

type Status struct {
	Name  string
	Value int
}

func (e Status) Status() Status {
	return e
}

type GetStatus interface {
	Status() Status
}

var (
	Null          = Status{"", -1}
	Indexed       = Status{"indexed", 0}
	Retrying      = Status{"retrying", 0}
	Queued        = Status{"queued", 1}
	Compared      = Status{"compared", 1}
	Downloading   = Status{"downloading", 2}
	Uploading     = Status{"uploading", 2}
	Skipped       = Status{"skipped", 3}
	FileExists    = Status{"file_exists", 3}
	Ignored       = Status{"ignored", 3}
	Complete      = Status{"complete", 4}
	FolderCreated = Status{"folder_created", 4}
	Canceled      = Status{"canceled", 5}
	Errored       = Status{"errored", 5}

	Included = []GetStatus{Indexed, Queued, Compared, Retrying, Downloading, Uploading, Complete, Canceled, Errored, FolderCreated}
	Excluded = []GetStatus{Skipped, Ignored, FileExists}
	Valid    = []GetStatus{Indexed, Queued, Compared, Retrying, Downloading, Uploading, Complete, FolderCreated}
	Invalid  = []GetStatus{Null, Canceled, Errored, Skipped, Ignored, FileExists}
	Running  = []GetStatus{Downloading, Uploading}
	Ended    = []GetStatus{Complete, Canceled, Errored, Skipped, Ignored, FileExists, FolderCreated}
)

func (e Status) String() string {
	return e.Name
}

func (e Status) Has(statuses ...GetStatus) bool {
	return e.Any(statuses...)
}

func (e Status) Is(statuses ...GetStatus) bool {
	return e.Any(statuses...)
}

func (e Status) IsNot(statuses ...GetStatus) bool {
	return !e.Any(statuses...)
}

func (e Status) is(status GetStatus) bool {
	return e.Name == status.Status().Name
}

func (e Status) Any(statuses ...GetStatus) bool {
	if len(statuses) == 0 {
		return true
	}
	for _, status := range statuses {
		if e.is(status) {
			return true
		}
	}
	return false
}

func SetStatus(old GetStatus, new GetStatus, err error) (Status, bool) {
	var setError bool
	if err != nil || new.Status().Is(Retrying) {
		setError = true
	}
	if old.Status().Is(Errored) && new.Status().Is(Running...) {
		new = old
	}

	return new.Status(), setError
}
