package logger

import "fmt"

func Info(s string) {
	fmt.Printf("%v %v\n", teal("[info]"), s)
}
func Sinfo(s string) {
	fmt.Printf("%v %v\n", teal("[info]"), s)
}

func Warn(s string) {
	fmt.Printf("%v %v\n", yellow("[warn]"), s)
}
func Swarn(s string) string {
	return fmt.Sprintf("%v %v\n", yellow("[warn]"), s)
}

func Err(s string) {
	fmt.Printf("%v %v\n", red("[err]"), s)
}
func Serr(s string) string {
	return fmt.Sprintf("%v %v\n", red("[err]"), s)
}

func Nomal(s string) {
	fmt.Printf("%v %v\n", green("[nomal]"), s)
}
func Snomal(s string) string {
	return fmt.Sprintf("%v %v\n", green("[nomal]"), s)
}

var (
	red    = color("\033[1;31m%s\033[0m")
	yellow = color("\033[1;33m%s\033[0m")
	teal   = color("\033[1;36m%s\033[0m")
	green  = color("\033[1;32m%s\033[0m")
)

func color(colorString string) func(...interface{}) string {
	sprint := func(args ...interface{}) string {
		return fmt.Sprintf(colorString,
			fmt.Sprint(args...))
	}
	return sprint
}
