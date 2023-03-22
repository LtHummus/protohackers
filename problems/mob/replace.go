package mob

import (
	"regexp"
	"strings"
)

// 7YWHMfk9JZe0LM0g1ZauHuiSxhI-HPVbees8gOSRTzOeroVi1op4tJNoiHr-1234
const (
	MobAddress = "7YWHMfk9JZe0LM0g1ZauHuiSxhI"
)

var (
	BoguscoinAddressRegex = regexp.MustCompile(`(^|)([A-Za-z0-9]{25,34})($| |\n)`)
	BogusCoinRawRegex     = regexp.MustCompile(`^7[A-Za-z0-9]{25,34}$`)
)

func swapAddress(line string) string {
	patchedWords := make([]string, 0)
	for _, curr := range strings.Split(line[:len(line)-1], " ") {
		patched := BogusCoinRawRegex.ReplaceAllString(curr, MobAddress)
		patchedWords = append(patchedWords, patched)
	}

	return strings.Join(patchedWords, " ") + "\n"
}
