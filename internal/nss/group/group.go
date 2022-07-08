package group

import (
	"context"
	"fmt"

	"github.com/ubuntu/aad-auth/internal/cache"
	"github.com/ubuntu/aad-auth/internal/logger"
	"github.com/ubuntu/aad-auth/internal/nss"
)

// Group is the nss group object.
type Group struct {
	name    string   /* username */
	passwd  string   /* user password */
	gid     uint     /* group ID */
	members []string /* Members of the group */
}

var testopts = []cache.Option{
	//cache.WithCacheDir("../cache"), cache.WithRootUid(1000), cache.WithRootGid(1000), cache.WithShadowGid(1000),
}

// NewByName returns a passwd entry from a name.
func NewByName(ctx context.Context, name string) (g Group, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to get group entry from name %q: %w", name, err)
		}
	}()

	logger.Debug(ctx, "Requesting a group entry matching name %q", name)

	if name == "shadow" {
		logger.Debug(ctx, "Ignoring shadow group as it's not in our database")
		return Group{}, nss.ErrNotFoundENoEnt
	}

	c, err := cache.New(ctx, testopts...)
	if err != nil {
		return Group{}, nss.ErrUnavailableENoEnt
	}
	defer c.Close()

	grp, err := c.GetGroupByName(ctx, name)
	if err != nil {
		// TODO: remove this wrapper and just print logs on error before converting to known format for the C lib.
		return Group{}, nss.ConvertErr(err)
	}

	return Group{
		name:    grp.Name,
		passwd:  grp.Password,
		gid:     uint(grp.GID),
		members: grp.Members,
	}, nil
}

// NewByGID returns a group entry from a GID.
func NewByGID(ctx context.Context, gid uint) (g Group, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to get group entry from GID %d: %w", gid, err)
		}
	}()

	logger.Debug(ctx, "Requesting an group entry matching GID %d", gid)

	c, err := cache.New(ctx, testopts...)
	if err != nil {

		return Group{}, nss.ErrUnavailableENoEnt
	}
	defer c.Close()

	grp, err := c.GetGroupByGID(ctx, gid)
	if err != nil {
		return Group{}, nss.ConvertErr(err)
	}

	return Group{
		name:    grp.Name,
		passwd:  grp.Password,
		gid:     uint(grp.GID),
		members: grp.Members,
	}, nil
}

var cacheIterateEntries *cache.Cache

// StartEntryIteration open a new cache for iteration.
func StartEntryIteration(ctx context.Context) error {
	c, err := cache.New(ctx, testopts...)
	if err != nil {
		// TODO: add context to error
		logger.Warn(ctx, "XXXXXXXXXXX: %v", err)
		return nss.ErrUnavailableENoEnt
	}
	cacheIterateEntries = c

	return nil
}

// EndEntryIteration closes the underlying DB.
func EndEntryIteration(ctx context.Context) error {
	if cacheIterateEntries == nil {
		logger.Warn(ctx, "group entry iteration ended without initialization first")
	}
	err := cacheIterateEntries.Close()
	cacheIterateEntries = nil
	return err
}

// NextEntry returns next available entry in Group. It will returns ENOENT from cache when the iteration is done.
func NextEntry(ctx context.Context) (g Group, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to get group entry: %w", err)
		}
	}()
	logger.Debug(ctx, "get next group entry")

	if cacheIterateEntries == nil {
		logger.Warn(ctx, "group entry iteration called without initialization first")
		return Group{}, nss.ErrUnavailableENoEnt
	}

	grp, err := cacheIterateEntries.NextGroupEntry(ctx)
	if err != nil {
		return Group{}, nss.ConvertErr(err)
	}

	return Group{
		name:    grp.Name,
		passwd:  grp.Password,
		gid:     uint(grp.GID),
		members: grp.Members,
	}, nil
}
