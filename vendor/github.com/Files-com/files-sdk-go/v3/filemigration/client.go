package file_migration

import (
	"context"
	"time"

	files_sdk "github.com/Files-com/files-sdk-go/v3"
	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type Client struct {
	files_sdk.Config
}

func (c *Client) Find(params files_sdk.FileMigrationFindParams, opts ...files_sdk.RequestResponseOption) (fileMigration files_sdk.FileMigration, err error) {
	err = files_sdk.Resource(c.Config, lib.Resource{Method: "GET", Path: "/file_migrations/{id}", Params: params, Entity: &fileMigration}, opts...)
	return
}

func Find(params files_sdk.FileMigrationFindParams, opts ...files_sdk.RequestResponseOption) (fileMigration files_sdk.FileMigration, err error) {
	return (&Client{}).Find(params, opts...)
}

func (c *Client) Wait(fileAction files_sdk.FileAction, status func(files_sdk.FileMigration), opts ...files_sdk.RequestResponseOption) (files_sdk.FileMigration, error) {
	var err error
	var migration files_sdk.FileMigration
	migration.Status = fileAction.Status
	migration.Id = fileAction.FileMigrationId
	if migration.Status == "completed" || migration.Status == "failed" || err != nil {
		return migration, nil
	}
	for {
		migration, err = c.Find(files_sdk.FileMigrationFindParams{Id: fileAction.FileMigrationId}, opts...)
		if err == nil {
			status(migration)
		}
		if migration.Status == "completed" || migration.Status == "failed" || err != nil {
			return migration, err
		}
		time.Sleep(time.Second * 1)
	}
}

func (c *Client) LogIterator(ctx context.Context, f files_sdk.FileMigration) files_sdk.IterI {
	return files_sdk.FilesMigrationLogIter{FileMigration: f, Context: ctx, Config: c.Config}.Init()
}
