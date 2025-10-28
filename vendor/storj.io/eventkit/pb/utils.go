package pb

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

func AsTimestamp(t time.Time) *Timestamp {
	return &Timestamp{
		Seconds: t.Unix(),
		Nanos:   int32(t.Nanosecond()),
	}
}

func (t *Timestamp) AsTime() time.Time {
	return time.Unix(t.Seconds, int64(t.Nanos)).UTC()
}

func (tag *Tag) KVString() string {
	return fmt.Sprintf("%s=%s", tag.Key, tag.ValueString())
}

func (tag *Tag) ValueString() string {
	if tag.Value == nil {
		return ""
	}
	switch t := tag.Value.(type) {
	default:
		panic(fmt.Sprintf("%T", tag.Value))
	case *Tag_String_:
		return string(t.String_)
	case *Tag_Int64:
		return fmt.Sprint(t.Int64)
	case *Tag_Double:
		return fmt.Sprint(t.Double)
	case *Tag_Bool:
		return fmt.Sprint(t.Bool)
	case *Tag_Bytes:
		return hex.EncodeToString(t.Bytes)
	case *Tag_DurationNs:
		return time.Duration(t.DurationNs).String()
	case *Tag_Timestamp:
		return t.Timestamp.AsTime().String()
	}
}

func (e *Event) TagsString() string {
	var parts []string
	for _, e := range e.Tags {
		parts = append(parts, e.KVString())
	}
	return strings.Join(parts, " ")
}
