package utils

import "github.com/fatih/color"

var colors = map[string]color.Attribute{
	"green":   color.FgGreen,
	"yellow":  color.FgYellow,
	"blue":    color.FgBlue,
	"magenta": color.FgMagenta,
	"cyan":    color.FgCyan,
	"red":     color.FgRed,
	"white":   color.FgWhite,
}

func Color(colorName string) *color.Color {
	return color.New(colors[colorName])
}

func DefaultColor() *color.Color {
	return Color("white")
}
