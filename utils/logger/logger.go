package logger

import "fmt"

func Info(t string, s ...string) {
	fmt.Printf("%v %v\n", teal(fmt.Sprintf("[%v]", t)), s)
}
func Sinfo(t string, s ...string) {
	fmt.Printf("%v %v\n", teal(fmt.Sprintf("[%v]", t)), s)
}

func Warn(t string, s ...string) {
	fmt.Printf("%v %v\n", yellow(fmt.Sprintf("[%v]", t)), s)
}
func Swarn(t string, s ...string) string {
	return fmt.Sprintf("%v %v\n", yellow(fmt.Sprintf("[%v]", t)), s)
}

func Err(t string, s ...string) {
	fmt.Printf("%v %v\n", red(fmt.Sprintf("[%v]", t)), s)
}
func Serr(t string, s ...string) string {
	return fmt.Sprintf("%v %v\n", red(fmt.Sprintf("[%v]", t)), s)
}

func Nomal(t string, s ...string) {
	fmt.Printf("%v %v\n", green(fmt.Sprintf("[%v]", t)), s)
}
func Snomal(t string, s ...string) string {
	return fmt.Sprintf("%v %v\n", green(fmt.Sprintf("[%v]", t)), s)
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
