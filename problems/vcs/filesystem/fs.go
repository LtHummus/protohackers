package filesystem

import (
	"errors"
	"path"
	"sort"
	"strings"
)

type Filesystem struct {
	Root *Dir
}

func NewFilesystem() *Filesystem {
	return &Filesystem{
		Root: &Dir{
			Name:    "/",
			Subdirs: []*Dir{},
			Files:   []*File{},
		},
	}
}

func checkFilename(name string) bool {
	return name[0] == '/'
}

func (f *Filesystem) getDir(dir string, create bool) *Dir {
	pathElements := strings.Split(dir, "/")

	currDir := f.Root

	for _, curr := range pathElements {
		nextDir := currDir.Subdir(curr)
		if nextDir == nil {
			if create {
				nextDir = &Dir{
					Name:    curr,
					Subdirs: []*Dir{},
					Files:   []*File{},
				}
				currDir.Subdirs = append(currDir.Subdirs, nextDir)
			} else {
				return nil
			}
		}
		currDir = nextDir
	}

	return currDir
}

func (f *Filesystem) List(dir string) (*ListResults, error) {
	if !checkFilename(dir) {
		return nil, errors.New("illegal dir name")
	}

	ret := &ListResults{
		Items: []ListItem{},
	}

	currDir := f.getDir(dir, false)

	if currDir == nil {
		return ret, nil
	}

	for _, curr := range currDir.Subdirs {
		ret.Items = append(ret.Items, ListItem{
			Name:     curr.Name,
			Revision: "DIR",
		})
	}

	for _, curr := range currDir.Files {
		ret.Items = append(ret.Items, ListItem{
			Name:     curr.Name,
			Revision: curr.LatestRevisionName(),
		})
	}

	sort.Slice(ret.Items, func(i, j int) bool {
		return ret.Items[i].Name < ret.Items[j].Name
	})

	return ret, nil
}

func (f *Filesystem) Put(name string, contents []byte) (string, error) {
	if !checkFilename(name) {
		return "", errors.New("illegal file name")
	}

	dirName, filename := path.Split(name)

	dir := f.getDir(dirName, true)
	return dir.PutFile(filename, contents)
}

func (f *Filesystem) Get(name string, revision string) ([]byte, error) {

	return nil, nil
}
