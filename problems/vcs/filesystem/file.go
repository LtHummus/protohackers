package filesystem

import "fmt"

type File struct {
	Name      string
	Revisions []*Revision
}

type Revision struct {
	Version  string
	Contents []byte
}

func (f *File) LatestRevisionName() string {
	return fmt.Sprintf("r%d", len(f.Revisions))
}

func (f *File) PutRevision(contents []byte) string {
	revisionName := fmt.Sprintf("r%d", len(f.Revisions)+1)
	f.Revisions = append(f.Revisions, &Revision{
		Version:  revisionName,
		Contents: contents,
	})

	return revisionName
}
