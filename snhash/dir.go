package snhash

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// HashDir builds a deterministic manifest for a directory tree.
func HashDir(root string, options ...Option) (Manifest, error) {
	opts := applyOptions(options)
	if opts.SourceName == "" {
		opts.SourceName = root
	}
	if opts.Canonicalization == defaultCanonicalization {
		opts.Canonicalization = "snhash-dir-v1:path-slash-sort-mode-size-digest"
	}

	var entries []Entry
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == root {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)

		entry := Entry{
			Path: rel,
			Mode: info.Mode().String(),
			Size: info.Size(),
		}
		switch {
		case d.Type()&os.ModeSymlink != 0:
			target, err := os.Readlink(path)
			if err != nil {
				return err
			}
			entry.Type = "symlink"
			linkManifest, err := HashBytes([]byte(target),
				WithAlgorithms(opts.Algorithms...),
				WithProfile(opts.Profile),
				WithMediaType("text/plain"),
				WithCanonicalization("symlink-target"),
				WithCreatedAt(opts.CreatedAt),
			)
			if err != nil {
				return err
			}
			entry.Size = int64(len(target))
			entry.Digests = linkManifest.Digests
		case d.IsDir():
			entry.Type = "directory"
			entry.Size = 0
		default:
			entry.Type = "file"
			fileManifest, err := HashFile(path,
				WithAlgorithms(opts.Algorithms...),
				WithProfile(opts.Profile),
				WithMediaType(defaultMediaType),
				WithCanonicalization(defaultCanonicalization),
				WithCreatedAt(opts.CreatedAt),
			)
			if err != nil {
				return err
			}
			entry.Digests = fileManifest.Digests
		}
		entries = append(entries, entry)
		return nil
	})
	if err != nil {
		return Manifest{}, err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Path < entries[j].Path })

	encoded := directoryCanonicalBytes(entries)
	rootManifest, err := HashBytes(encoded,
		WithAlgorithms(opts.Algorithms...),
		WithProfile(opts.Profile),
		WithMediaType("application/vnd.snhash.directory+json"),
		WithCanonicalization(opts.Canonicalization),
		WithSourceName(opts.SourceName),
		WithCreatedAt(opts.CreatedAt),
	)
	if err != nil {
		return Manifest{}, err
	}
	rootManifest.Size = int64(len(encoded))
	rootManifest.Entries = entries
	rootManifest.Custody = opts.Custody
	return rootManifest, nil
}

func directoryCanonicalBytes(entries []Entry) []byte {
	var b strings.Builder
	for _, entry := range entries {
		b.WriteString(entry.Path)
		b.WriteByte('\x00')
		b.WriteString(entry.Type)
		b.WriteByte('\x00')
		b.WriteString(entry.Mode)
		b.WriteByte('\x00')
		for _, digest := range normalizeDigests(entry.Digests) {
			b.WriteString(digest.Wire())
			b.WriteByte('\x00')
		}
		b.WriteByte('\n')
	}
	return []byte(b.String())
}
