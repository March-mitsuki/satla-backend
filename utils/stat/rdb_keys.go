package stat

import "fmt"

func MakeRdbKey(wsroom string) string {
	return fmt.Sprintf("roomState:%v", wsroom)
}
