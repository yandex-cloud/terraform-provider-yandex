package helpers

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
	"github.com/ydb-platform/ydb-go-sdk/v3/scheme"
)

type schemeClient interface {
	Database() string
	ListDirectory(ctx context.Context, path string) (d scheme.Directory, err error)
}

func IsDirectoryExists(ctx context.Context, c schemeClient, directory string) (
	exists bool, _ error,
) {
	if !strings.HasPrefix(directory, c.Database()) {
		return false, xerrors.WithStackTrace(fmt.Errorf(
			"path '%s' must be inside database '%s'",
			directory, c.Database(),
		))
	}
	if directory == c.Database() {
		return true, nil
	}
	parentDirectory, childDirectory := path.Split(directory)
	parentDirectory = strings.TrimRight(parentDirectory, "/")

	if exists, err := IsDirectoryExists(ctx, c, parentDirectory); err != nil {
		return false, xerrors.WithStackTrace(err)
	} else if !exists {
		return false, nil
	}

	d, err := c.ListDirectory(ctx, parentDirectory)
	if err != nil {
		return false, xerrors.WithStackTrace(err)
	}
	for i := range d.Children {
		if d.Children[i].Name != childDirectory {
			continue
		}
		if t := d.Children[i].Type; t != scheme.EntryDirectory {
			return false, xerrors.WithStackTrace(fmt.Errorf(
				"entry '%s' in path '%s' is not a directory: %s",
				childDirectory, parentDirectory, t.String(),
			))
		}

		return true, nil
	}

	return false, nil
}

func IsEntryExists(ctx context.Context, c schemeClient, absPath string, entryTypes ...scheme.EntryType) (
	exists bool, _ error,
) {
	if !strings.HasPrefix(absPath, c.Database()) {
		return false, xerrors.WithStackTrace(fmt.Errorf(
			"entry path '%s' must be inside database '%s'",
			absPath, c.Database(),
		))
	} else if absPath == c.Database() {
		return false, xerrors.WithStackTrace(fmt.Errorf(
			"entry path '%s' cannot be equals database name '%s'",
			absPath, c.Database(),
		))
	}
	directory, entryName := path.Split(absPath)
	if exists, err := IsDirectoryExists(ctx, c, strings.TrimRight(directory, "/")); err != nil {
		return false, xerrors.WithStackTrace(err)
	} else if !exists {
		return false, nil
	}
	d, err := c.ListDirectory(ctx, directory)
	if err != nil {
		return false, xerrors.WithStackTrace(fmt.Errorf(
			"list directory '%s' failed: %w",
			directory, err,
		))
	}
	for i := range d.Children {
		if d.Children[i].Name != entryName {
			continue
		}
		childrenType := d.Children[i].Type
		for _, entryType := range entryTypes {
			if childrenType == entryType {
				return true, nil
			}
		}

		return false, xerrors.WithStackTrace(fmt.Errorf(
			"entry type of '%s' (%s) in path '%s' is not corresponds to %v",
			entryName, childrenType, directory, entryTypes,
		))
	}

	return false, nil
}
