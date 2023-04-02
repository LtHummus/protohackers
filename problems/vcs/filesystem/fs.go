package filesystem

import (
	"errors"
	"fmt"
	"sort"
	"sync"
)

type Filesystem struct {
	Root *Dir

	lock sync.Mutex
}

func NewFilesystem() *Filesystem {
	return &Filesystem{
		Root: newDir(""),
	}
}

func checkFilename(name string) bool {
	return name[0] == '/'
}

func (f *Filesystem) getDir(path []string, create bool) *Dir {
	currDir := f.Root

	for _, curr := range path {
		nextDir := currDir.Subdirs[curr]
		if nextDir == nil && !create {
			return nil
		} else if nextDir == nil && create {
			nextDir = newDir(curr)
			currDir.Subdirs[curr] = nextDir
		}
		currDir = nextDir
	}

	return currDir
}

func (f *Filesystem) List(dir string) (*ListResults, error) {
	f.lock.Lock()
	defer f.lock.Unlock()

	path, err := parseFilepath(dir)
	if err != nil {
		return nil, err
	}

	ret := &ListResults{
		Items: []ListItem{},
	}

	currDir := f.getDir(path, false)

	if currDir == nil {
		return ret, nil
	}

	for _, curr := range currDir.Subdirs {
		ret.Items = append(ret.Items, ListItem{
			Name:     fmt.Sprintf("%s/", curr.Name),
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

	f.lock.Lock()
	defer f.lock.Unlock()

	path, err := parseFilepath(name)
	if err != nil {
		return "", err
	}

	dir := f.getDir(path[:len(path)-1], true)
	return dir.PutFile(path[len(path)-1], contents)
}

func (f *Filesystem) Get(name string, revision string) ([]byte, error) {
	if !checkFilename(name) {
		return nil, errors.New("illegal file name")
	}

	f.lock.Lock()
	defer f.lock.Unlock()

	path, err := parseFilepath(name)
	if err != nil {
		return nil, err
	}

	dir := f.getDir(path[:len(path)-1], false)
	if dir == nil {
		return nil, errors.New("no such file")
	}

	file := dir.Files[path[len(path)-1]]
	if file == nil {
		return nil, errors.New("no such file")
	}

	return file.GetRevision(revision)
}
