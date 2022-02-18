package util

import (
	"nathanman/config"

	"github.com/TwiN/go-color"
)

func LogInYellow(in string) {
	config.Logger.Println(color.InYellow(in))
}

func PanicInRed(in string) {
	config.Logger.Panicln(color.InRed(in))
}
