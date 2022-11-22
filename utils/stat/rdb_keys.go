package stat

import "fmt"

func MakeAutoRdbKey(wsroom string) string {
	return fmt.Sprintf("%v:autoState", wsroom)
}

func MakeSubtitleRdbKey(wsroom string) string {
	return fmt.Sprintf("%v:subtitleState", wsroom)
}

func MakeStyleRdbKey(wsroom string) string {
	return fmt.Sprintf("%v:styleState", wsroom)
}
