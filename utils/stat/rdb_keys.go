package stat

import "fmt"

func MakeRdbKeys(wsroom string) string {
	return fmt.Sprintf("roomState:%v", wsroom)
}
