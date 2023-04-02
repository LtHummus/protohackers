package filesystem

type Dir struct {
	Name string

	Subdirs []*Dir
	Files   []*File
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
	currFile := d.File(name)
	if currFile != nil {
		r := currFile.PutRevision(contents)
		return r, nil
	}

	f := &File{
		Name:      name,
		Revisions: []*Revision{},
	}

	d.Files = append(d.Files, f)

	return f.PutRevision(contents), nil
}
