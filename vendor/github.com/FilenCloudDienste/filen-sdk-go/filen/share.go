package filen

import (
	"context"
	"crypto/rsa"
	"fmt"
	"github.com/FilenCloudDienste/filen-sdk-go/filen/client"
	"github.com/FilenCloudDienste/filen-sdk-go/filen/crypto"
	"github.com/FilenCloudDienste/filen-sdk-go/filen/types"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

// renameSharedItem updates the metadata of an item that is shared with another user.
// It encrypts the metadata with the recipient's public key to maintain end-to-end encryption.
func (api *Filen) renameSharedItem(ctx context.Context, item types.FileSystemObject, receiverId int, metadata string, key rsa.PublicKey) error {
	encryptedMeta, err := crypto.PublicEncrypt(&key, metadata)
	if err != nil {
		return err
	}

	return api.Client.PostV3ItemSharedRename(ctx, item.GetUUID(), receiverId, encryptedMeta)
}

// renameLinkedItem updates the metadata of an item that has a public link.
// It encrypts the metadata with the link's encryption key to maintain security.
func (api *Filen) renameLinkedItem(ctx context.Context, item types.FileSystemObject, linkUUID string, encryptedMeta crypto.EncryptedString) error {
	err := api.Client.PostV3ItemLinkedRename(ctx, item.GetUUID(), linkUUID, encryptedMeta)
	if err != nil {
		return err
	}
	return nil
}

// addItemToDirectoryPublicLink adds an item to an existing directory public link.
// This is used when items are added to a directory that is already publicly shared.
func (api *Filen) addItemToDirectoryPublicLink(ctx context.Context, uuid, parentUUID, itemType, linkUUID string, encryptedMeta crypto.EncryptedString, linkKey crypto.EncryptedString) error {
	return api.Client.PostV3DirLinkAdd(ctx, client.V3DirLinkAddRequest{
		UUID:       uuid,
		ParentUUID: parentUUID,
		LinkUUID:   linkUUID,
		ItemType:   itemType,
		Metadata:   encryptedMeta,
		LinkKey:    linkKey,
		Expiration: "never",
	})
}

// updateMaybeSharedItem updates the metadata for an item in all its shared contexts.
// This ensures that when an item is renamed or modified, the changes are visible
// to all users and public links that have access to it.
func (api *Filen) updateMaybeSharedItem(ctx context.Context, item types.NonRootFileSystemObject) error {
	g, gCtx := errgroup.WithContext(ctx)

	var sharedResult *client.V3ItemSharedResponse
	var linkedResult *client.V3ItemLinkedResponse

	g.Go(func() error {
		var err error
		sharedResult, err = api.Client.PostV3ItemShared(gCtx, item.GetUUID())
		return err
	})

	g.Go(func() error {
		var err error
		linkedResult, err = api.Client.PostV3ItemLinked(gCtx, item.GetUUID())
		return err
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("get shared or linked status: %w", err)
	}

	g, gCtx = errgroup.WithContext(ctx)
	g.SetLimit(MaxSmallCallers)
	metaData, err := item.GetMeta(api.FileEncryptionVersion)
	if err != nil {
		return fmt.Errorf("get meta: %w", err)
	}
	for _, user := range sharedResult.Users {
		g.Go(func() error {
			publicKey, err := crypto.PublicKeyFromString(user.PublicKey)
			if err != nil {
				return fmt.Errorf("parse public key: %w", err)
			}
			return api.renameSharedItem(gCtx, item, user.ID, metaData, *publicKey)
		})
	}
	for _, link := range linkedResult.Links {
		g.Go(func() error {
			keyStr, err := api.DecryptMeta(link.Key)
			if err != nil {
				return fmt.Errorf("decrypt key: %w", err)
			}
			key, err := api.GetMetaCrypterFromKeyString(keyStr, -1)
			if err != nil {
				return fmt.Errorf("make key: %w", err)
			}
			return api.renameLinkedItem(gCtx, item, link.LinkUUID, key.EncryptMeta(metaData))
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("rename shared or linked item: %w", err)
	}

	return nil
}

// shareData represents the metadata needed to share an item.
// It is used internally when sharing items with users or through public links.
type shareData struct {
	UUID       string // The UUID of the item
	ParentUUID string // The UUID of the parent directory
	Metadata   string // The item's metadata in plaintext
	Type       string // The type of the item ("file" or "folder")
}

// updateItemWithMaybeSharedParent updates the shared/linked status of an item
// if its parent is shared or linked. This ensures that newly created or moved
// items inherit the sharing properties of their parent directory.
//
// This function needs to be called whenever an item is moved, created or renamed.
func (api *Filen) updateItemWithMaybeSharedParent(ctx context.Context, item types.NonRootFileSystemObject) error {
	parentUUID := item.GetParent()
	g, gCtx := errgroup.WithContext(ctx)
	var sharedResult *client.V3ItemSharedResponse
	var linkedResult *client.V3DirLinkedResponse

	g.Go(func() error {
		var err error
		sharedResult, err = api.Client.PostV3ItemShared(gCtx, parentUUID)
		return err
	})

	g.Go(func() error {
		var err error
		linkedResult, err = api.Client.PostV3DirLinked(gCtx, parentUUID)
		return err
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("get parent shared or linked status: %w", err)
	}

	if !sharedResult.Shared && !linkedResult.Linked {
		return nil
	}

	dataToShare := make([]shareData, 0, 1)

	meta, err := item.GetMeta(api.FileEncryptionVersion)
	if err != nil {
		return fmt.Errorf("get meta: %w", err)
	}
	dataToShare = append(dataToShare, shareData{
		UUID:       item.GetUUID(),
		ParentUUID: item.GetParent(),
		Metadata:   meta,
		Type:       "",
	})

	if dir, ok := item.(*types.Directory); ok {
		dataToShare[0].Type = "folder"
		files, dirs, err := api.ListRecursive(ctx, dir)
		if err != nil {
			return fmt.Errorf("list recursive: %w", err)
		}
		for _, file := range files {
			meta, err := file.GetMeta(api.FileEncryptionVersion)
			if err != nil {
				return fmt.Errorf("get meta: %w", err)
			}
			dataToShare = append(dataToShare, shareData{
				UUID:       file.UUID,
				ParentUUID: file.ParentUUID,
				Metadata:   meta,
				Type:       "file",
			})
		}
		for _, dir := range dirs {
			meta, err := dir.GetMeta(api.FileEncryptionVersion)
			if err != nil {
				return fmt.Errorf("get meta: %w", err)
			}
			dataToShare = append(dataToShare, shareData{
				UUID:       dir.UUID,
				ParentUUID: dir.ParentUUID,
				Metadata:   meta,
				Type:       "folder",
			})
		}
	} else {
		dataToShare[0].Type = "file"
	}

	g, gCtx = errgroup.WithContext(ctx)
	g.SetLimit(MaxSmallCallers)

	for _, user := range sharedResult.Users {
		key, err := crypto.PublicKeyFromString(user.PublicKey)
		if err != nil {
			return fmt.Errorf("parse public key: %w", err)
		}
		for _, data := range dataToShare {
			encrypted, err := crypto.PublicEncrypt(key, data.Metadata)
			if err != nil {
				return fmt.Errorf("public encrypt: %w", err)
			}
			g.Go(func() error {
				_ = item
				return api.Client.PostV3ItemShare(gCtx, client.V3ItemShareRequest{
					UUID:       data.UUID,
					ParentUUID: data.ParentUUID,
					Email:      user.Email,
					Type:       data.Type,
					Metadata:   encrypted,
				})
			})
		}
	}

	for _, link := range linkedResult.Links {
		keyStr, err := api.DecryptMeta(link.Key)
		if err != nil {
			return fmt.Errorf("decrypt meta: %w", err)
		}
		key, err := api.GetMetaCrypterFromKeyString(keyStr, -1)
		if err != nil {
			return fmt.Errorf("make key: %w", err)
		}
		for _, data := range dataToShare {
			g.Go(func() error {
				return api.Client.PostV3DirLinkAdd(gCtx, client.V3DirLinkAddRequest{
					UUID:       data.UUID,
					ParentUUID: data.ParentUUID,
					LinkUUID:   link.UUID,
					ItemType:   data.Type,
					Metadata:   key.EncryptMeta(data.Metadata),
					LinkKey:    link.Key,
					Expiration: "never",
				})
			})
		}
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("share or link: %w", err)
	}
	return nil
}

// publicLinkFile creates a public link for a file.
// Returns the link UUID which can be used to construct a shareable URL.
func (api *Filen) publicLinkFile(ctx context.Context, file types.File) (string, error) {
	return api.Client.PostV3FileLinkEditEnable(ctx, file)
}

// publicLinkDir creates a public link for a directory and all its contents.
// This ensures that the entire directory tree is accessible through the link.
// Returns the link UUID which can be used to construct a shareable URL.
func (api *Filen) publicLinkDir(ctx context.Context, dir *types.Directory) (string, error) {
	linkUUID := uuid.NewString()
	key := crypto.GenerateRandomString(32)
	files, dirs, err := api.ListRecursive(ctx, dir)
	linkKeyEncrypted := api.EncryptMeta(key)
	if err != nil {
		return "", fmt.Errorf("list recursive: %w", err)
	}

	g, gCtx := errgroup.WithContext(ctx)
	g.SetLimit(MaxSmallCallers)

	g.Go(func() error {
		meta, err := dir.GetMeta(api.FileEncryptionVersion)
		if err != nil {
			return fmt.Errorf("get meta: %w", err)
		}
		return api.Client.PostV3DirLinkAdd(gCtx, client.V3DirLinkAddRequest{
			UUID:       dir.GetUUID(),
			ParentUUID: "base",
			LinkUUID:   linkUUID,
			ItemType:   "folder",
			Metadata:   api.EncryptMeta(meta),
			LinkKey:    linkKeyEncrypted,
			Expiration: "never",
		})
	})

	for _, file := range files {
		g.Go(func() error {
			meta, err := file.GetMeta(api.FileEncryptionVersion)
			if err != nil {
				return fmt.Errorf("get meta: %w", err)
			}
			return api.Client.PostV3DirLinkAdd(gCtx, client.V3DirLinkAddRequest{
				UUID:       file.GetUUID(),
				ParentUUID: file.GetParent(),
				LinkUUID:   linkUUID,
				ItemType:   "file",
				Metadata:   api.EncryptMeta(meta),
				LinkKey:    linkKeyEncrypted,
				Expiration: "never",
			})
		})
	}
	for _, dir := range dirs {
		g.Go(func() error {
			meta, err := dir.GetMeta(api.FileEncryptionVersion)
			if err != nil {
				return fmt.Errorf("get meta: %w", err)
			}
			return api.Client.PostV3DirLinkAdd(gCtx, client.V3DirLinkAddRequest{
				UUID:       dir.GetUUID(),
				ParentUUID: dir.GetParent(),
				LinkUUID:   linkUUID,
				ItemType:   "folder",
				Metadata:   api.EncryptMeta(meta),
				LinkKey:    linkKeyEncrypted,
				Expiration: "never",
			})
		})
	}

	if err := g.Wait(); err != nil {
		return "", fmt.Errorf("share or link: %w", err)
	}
	return linkUUID, nil
}

// PublicLinkItem creates a public link for a file or directory.
// This link can be shared with anyone, even those without Filen accounts.
// Returns the LinkUUID for the link, which can be used to construct a shareable URL.
func (api *Filen) PublicLinkItem(ctx context.Context, item types.NonRootFileSystemObject) (string, error) {
	if dir, ok := item.(*types.Directory); ok {
		return api.publicLinkDir(ctx, dir)
	} else if file, ok := item.(*types.File); ok {
		return api.publicLinkFile(ctx, *file)
	}
	return "", fmt.Errorf("unknown type: %T", item)
}

// shareItemToUserNonRecursive shares a single item with another Filen user.
// It encrypts the item's metadata with the recipient's public key to maintain end-to-end encryption.
// This function does not share child items if the item is a directory.
func (api *Filen) shareItemToUserNonRecursiveWithParent(ctx context.Context, item types.NonRootFileSystemObject, parentString string, email string, key *rsa.PublicKey) error {
	metaStr, err := item.GetMeta(api.FileEncryptionVersion)
	if err != nil {
		return fmt.Errorf("get meta: %w", err)
	}
	meta, err := crypto.PublicEncrypt(key, metaStr)
	if err != nil {
		return fmt.Errorf("encrypt meta: %w", err)
	}

	var itemType string
	if _, ok := item.(*types.File); ok {
		itemType = "file"
	} else if _, ok := item.(*types.Directory); ok {
		itemType = "folder"
	} else {
		return fmt.Errorf("unknown type: %T", item)
	}

	err = api.Client.PostV3ItemShare(ctx, client.V3ItemShareRequest{
		UUID:       item.GetUUID(),
		ParentUUID: parentString,
		Email:      email,
		Type:       itemType,
		Metadata:   meta,
	})

	if err != nil {
		return fmt.Errorf("share: %w when sharing %s", err, item.GetName())
	}
	return nil
}

// shareDirToUser shares a directory and all its contents with another Filen user.
// This ensures that the entire directory tree is accessible to the recipient.
func (api *Filen) shareDirToUser(ctx context.Context, dir *types.Directory, email string, key *rsa.PublicKey) error {
	files, dirs, err := api.ListRecursive(ctx, dir)
	if err != nil {
		return fmt.Errorf("list recursive: %w", err)
	}

	g, gCtx := errgroup.WithContext(ctx)
	g.SetLimit(MaxSmallCallers)

	g.Go(func() error {
		return api.shareItemToUserNonRecursiveWithParent(gCtx, dir, "none", email, key)
	})

	for _, file := range files {
		g.Go(func() error {
			return api.shareItemToUserNonRecursiveWithParent(gCtx, file, file.GetParent(), email, key)
		})
	}
	for _, dir := range dirs {
		g.Go(func() error {
			return api.shareItemToUserNonRecursiveWithParent(gCtx, dir, dir.GetParent(), email, key)
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("share or link: %w", err)
	}
	return nil
}

// ShareItemToUser shares a file or directory with another Filen user.
// If the item is a directory, all its contents are shared recursively.
// This function handles fetching the recipient's public key and encrypting
// the metadata accordingly to maintain end-to-end encryption.
func (api *Filen) ShareItemToUser(ctx context.Context, item types.NonRootFileSystemObject, email string) error {
	publicKeyObj, err := api.Client.PostV3UserPublicKey(ctx, email)
	if err != nil {
		return fmt.Errorf("get public key: %w", err)
	}
	publicKey, err := crypto.PublicKeyFromString(publicKeyObj.PublicKey)
	if err != nil {
		return fmt.Errorf("parse public key: %w", err)
	}

	if dir, ok := item.(*types.Directory); ok {
		return api.shareDirToUser(ctx, dir, email, publicKey)
	} else if file, ok := item.(*types.File); ok {
		return api.shareItemToUserNonRecursiveWithParent(ctx, file, "none", email, publicKey)
	}
	return fmt.Errorf("unknown type: %T", item)

}

// IsItemShared checks if an item is shared with other Filen users.
// Returns true if the item is currently shared with at least one user.
func (api *Filen) IsItemShared(ctx context.Context, item types.NonRootFileSystemObject) (bool, error) {
	resp, err := api.Client.PostV3ItemShared(ctx, item.GetUUID())
	if err != nil {
		return false, fmt.Errorf("get shared status: %w", err)
	}
	return resp.Shared, nil
}

// IsItemLinked checks if an item has a public link.
// Returns true if the item currently has at least one public link.
func (api *Filen) IsItemLinked(ctx context.Context, item types.NonRootFileSystemObject) (bool, error) {
	resp, err := api.Client.PostV3ItemLinked(ctx, item.GetUUID())
	if err != nil {
		return false, fmt.Errorf("get linked status: %w", err)
	}
	return resp.Linked, nil
}
