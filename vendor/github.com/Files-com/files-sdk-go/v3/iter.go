package files_sdk

import (
	"github.com/Files-com/files-sdk-go/v3/lib"
)

type ListParams struct {
	PerPage  int64  `json:"per_page,omitempty" url:"per_page,omitempty" required:"false"`
	Cursor   string `json:"cursor,omitempty" url:"cursor,omitempty" required:"false"`
	MaxPages int64  `json:"-" url:"-"`
}

// ListParamsContainer is a general interface for which all list parameter
// structs should comply. They achieve this by embedding a ListParams struct
// and inheriting its implementation of this interface.
type ListParamsContainer interface {
	GetListParams() *ListParams
}

// GetListParams returns a ListParams struct (itself). It exists because any
// structs that embed ListParams will inherit it, and thus implement the
// ListParamsContainer interface.
func (p *ListParams) GetListParams() *ListParams {
	return p
}

type OnPageError func(error) (*[]interface{}, error)
type Query func(params lib.Values, opts ...RequestResponseOption) (*[]interface{}, string, error)

type IterI interface {
	Next() bool
	Current() interface{}
	Err() error
}

var _ IterI = (*Iter)(nil)

type TypedIterI[T any] interface {
	Next() bool
	Current() interface{}
	Resource() T
	Err() error
}

type IterPagingI interface {
	IterI
	EOFPage() bool
}

var _ IterPagingI = (*Iter)(nil)

type ResourceIterator interface {
	Iterate(interface{}, ...RequestResponseOption) (IterI, error)
}

type ReloadIterator interface {
	Reload(opts ...RequestResponseOption) IterI
}

var _ ReloadIterator = (*Iter)(nil)

type ResourceLoader interface {
	LoadResource(interface{}, ...RequestResponseOption) (interface{}, error)
}

type Identifier interface {
	Identifier() interface{}
}

type Iterable interface {
	Iterable() bool
}

type Iter struct {
	Query
	ListParams   ListParamsContainer
	Params       []interface{}
	CurrentIndex int
	Page         int64
	Values       *[]interface{}
	Cursor       string
	Error        error
	OnPageError
	requestResponseOptions []RequestResponseOption
}

// Err returns the error, if any,
// that caused the Iter to stop.
// It must be inspected
// after Next returns false.
func (i *Iter) Err() error {
	return i.Error
}

func (i *Iter) Current() interface{} {
	return (*i.Values)[i.CurrentIndex]
}

func (i *Iter) GetParams() *ListParams {
	return i.ListParams.GetListParams()
}

func (i *Iter) ExportParams() (lib.ExportValues, error) {
	p := lib.Params{Params: i.GetParams()}
	paramValues, err := p.ToValues()
	if err != nil {
		return lib.ExportValues{}, err
	}
	listParamValues, err := lib.Params{Params: i.ListParams}.ToValues()

	if err != nil {
		return lib.ExportValues{}, err
	}

	for key, value := range paramValues {
		listParamValues.Set(key, value[0])
	}

	return lib.ExportValues{Values: listParamValues}, nil
}

func (i *Iter) GetPage() bool {
	if i.GetParams().MaxPages != 0 && i.Page >= i.GetParams().MaxPages {
		return false
	}

	i.CurrentIndex = 0
	i.Page += 1

	if i.Values != nil && i.Cursor == "" {
		return false
	}
	params, _ := i.ExportParams()
	i.Values, i.Cursor, i.Error = i.Query(params, i.requestResponseOptions...)
	i.SetCursor(i.Cursor)
	if i.Error != nil && i.OnPageError != nil {
		i.Values, i.Error = i.OnPageError(i.Error)
	}
	return i.Error == nil && len(*i.Values) != 0
}

func (i *Iter) EOFPage() bool {
	return len(*i.Values) == i.CurrentIndex+1
}

func (i *Iter) Paging() bool {
	return true
}

func (i *Iter) GetCursor() string {
	return i.GetParams().Cursor
}

func (i *Iter) SetCursor(cursor string) {
	i.GetParams().Cursor = cursor
	i.Cursor = cursor
}

// Next iterates the results in i.Current() or i.`ResourceName`().
// It returns true until there are no results remaining.
// To adjust the number of results set ListParams.PerPage.
// To have it auto-paginate set ListParams.MaxPages, default is 1.
//
// To iterate over all results use the following pattern.
//
//	for i.Next() {
//	  i.Current()
//	}
func (i *Iter) Next() bool {
	if i.Values == nil {
		return i.GetPage() && len(*i.Values) > 0
	} else if len(*i.Values) > i.CurrentIndex+1 {
		i.CurrentIndex += 1
		return true
	}

	if i.EOFPage() {
		return i.GetPage()
	}

	return false
}

func (i *Iter) NextPage() bool {
	return i.Cursor != ""
}

// Reload ignores any id passed in and creates a new reset Iter
func (i *Iter) Reload(opts ...RequestResponseOption) IterI {
	newIter := *i
	newIter.ListParams = &ListParams{}
	newIter.requestResponseOptions = opts
	return &newIter
}
