package filesystem

import (
	"fmt"
	"strings"
)

type ListItem struct {
	Name     string
	Revision string
}

func (li *ListItem) String() string {
	return fmt.Sprintf("%s %s", li.Name, li.Revision)
}

type ListResults struct {
	Items []ListItem
}

func (lr *ListResults) String() string {
	b := strings.Builder{}

	b.WriteString(fmt.Sprintf("OK %d\n", len(lr.Items)))
	for _, curr := range lr.Items {
		b.WriteString(fmt.Sprintf("%s\n", curr.String()))
	}
	return b.String()
}
