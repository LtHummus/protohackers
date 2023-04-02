package filesystem

import "errors"

type Dir struct {
	Name string

	Subdirs map[string]*Dir
	Files   map[string]*File
}

func newDir(name string) *Dir {
	return &Dir{
		Name:    name,
		Subdirs: map[string]*Dir{},
		Files:   map[string]*File{},
	}
}

func (d *Dir) Subdir(name string) *Dir {
	for _, curr := range d.Subdirs {
		if curr.Name == name {
			return curr
		}
	}

	return nil
}

func (d *Dir) File(name string) *File {
	for _, curr := range d.Files {
		if curr.Name == name {
			return curr
		}
	}

	return nil
}

func (d *Dir) PutFile(name string, contents []byte) (string, error) {
	for _, c := range contents {
		if c < 9 || c > 13 && c < 32 || c > 127 {
			return "", errors.New("text is binary")
		}
	}
	currFile := d.File(name)
	if currFile != nil {
		r := currFile.PutRevision(contents)
		return r, nil
	}

	f := &File{
		Name:      name,
		Revisions: []*Revision{},
	}

	d.Files[name] = f

	return f.PutRevision(contents), nil
}
