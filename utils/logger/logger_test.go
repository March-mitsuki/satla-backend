package logger

import (
	"fmt"
	"testing"
)

func TestErr(t *testing.T) {
	d := "some err msg here"
	s := Serr(d)
	fmt.Println(s)
}

func TestNomal(t *testing.T) {
	d := "some nomal msg here"
	s := Snomal(d)
	fmt.Println(s)
}
