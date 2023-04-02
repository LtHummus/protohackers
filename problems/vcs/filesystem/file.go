package filesystem

import (
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"reflect"
	"strconv"
	"strings"
)

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
	if len(f.Revisions) != 0 {
		lastRevision := f.Revisions[len(f.Revisions)-1]
		if reflect.DeepEqual(lastRevision.Contents, contents) {
			log.Warn().Msg("contents the same, not making new revision")
			return fmt.Sprintf("r%d", len(f.Revisions))
		}
	}

	revisionName := fmt.Sprintf("r%d", len(f.Revisions)+1)
	f.Revisions = append(f.Revisions, &Revision{
		Version:  revisionName,
		Contents: contents,
	})

	return revisionName
}

func (f *File) GetRevision(revision string) ([]byte, error) {
	var revisionNum int
	var err error
	if revision != "" {
		strippedRevision := strings.TrimPrefix(revision, "r")
		revisionNum, err = strconv.Atoi(strippedRevision)
	} else {
		revisionNum = len(f.Revisions)
	}
	if err != nil {
		return nil, errors.New("no such revision")
	}

	if revisionNum <= 0 || revisionNum > len(f.Revisions) {
		return nil, errors.New("no such revision")
	}

	return f.Revisions[revisionNum-1].Contents, nil
}
