package themes

import (
	"io/fs"

	"golang.org/x/exp/slices"
)

func sortDirsFirst(less func(i, j fs.FileInfo) bool, entries []fs.FileInfo) {
	slices.SortFunc(entries, func(i, j fs.FileInfo) bool {
		if i.IsDir() {
			if j.IsDir() {
				return less(i, j)
			}

			return true
		} else if j.IsDir() {
			return false
		}

		return less(i, j)
	})
}

const (
	sortSize     = "size"
	sortSizeDesc = "size_desc"
	sortTime     = "time"
	sortTimeDesc = "time_desc"
)

func sortBy(param string, entries []fs.FileInfo) (dirs, files []fs.FileInfo) {
	var less func(i, j fs.FileInfo) bool
	switch param {
	case sortSize:
		less = func(i, j fs.FileInfo) bool {
			if i.IsDir() {
				return i.Name() < j.Name()
			}

			return i.Size() < j.Size()
		}
	case sortSizeDesc:
		less = func(i, j fs.FileInfo) bool {
			if i.IsDir() {
				return i.Name() < j.Name()
			}

			return i.Size() > j.Size()
		}
	case sortTime:
		less = func(i, j fs.FileInfo) bool {
			return i.ModTime().Before(j.ModTime())
		}
	case sortTimeDesc:
		less = func(i, j fs.FileInfo) bool {
			return i.ModTime().After(j.ModTime())
		}
	default:
		less = func(i, j fs.FileInfo) bool {
			return i.Name() < j.Name()
		}
	}

	sortDirsFirst(less, entries)

	di := slices.IndexFunc(entries, func(fi fs.FileInfo) bool { return !fi.IsDir() })

	return entries[:di], entries[di:]
}

const (
	paramSort = "sortBy"
)
