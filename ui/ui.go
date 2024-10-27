package ui

import (
	"github.com/pterm/pterm"
	"github.com/pterm/pterm/putils"
)

func Update(title string, time string) {
	a, _ := pterm.DefaultArea.WithFullscreen().WithCenter().Start()
	s, _ := pterm.DefaultBigText.WithLetters(putils.LettersFromString(title)).Srender()
	t, _ := pterm.DefaultBigText.WithLetters(putils.LettersFromString(time)).Srender()

	a.Update(s + "\n" + t)
}
