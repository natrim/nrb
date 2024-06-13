package lib

import (
	"fmt"
	"os"
)

const (
	ColorBlack   = "\033[0;30m"
	ColorRed     = "\033[0;31m"
	ColorGreen   = "\033[0;32m"
	ColorYellow  = "\033[0;33m"
	ColorBlue    = "\033[0;34m"
	ColorMagenta = "\033[0;35m"
	ColortCyan   = "\033[0;36m"
	ColortWhite  = "\033[0;37m"
	ColorClear   = "\033[0m"
)

var ERR = "× ERR:"
var INFO = ">"
var OK = "✓"
var RELOAD = "↻"
var ITEM = "-"
var DASH = "–"

func UseColor(use bool) {
	if use {
		ERR = Red(ERR)
		INFO = Yellow(INFO)
		OK = Green(OK)
		RELOAD = Blue(RELOAD)
		ITEM = Magenta(ITEM)
		DASH = Blue(DASH)
	}
}

func Print(a ...any) {
	_, _ = fmt.Fprintln(os.Stdout, a...)
}

func Printf(format string, a ...any) {
	_, _ = fmt.Fprintf(os.Stdout, format, a...)
}

func Printe(a ...any) {
	_, _ = fmt.Fprintln(os.Stderr, a...)
}

func Printef(format string, a ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format, a...)
}

func PrintError(a ...any) {
	_, _ = fmt.Fprintln(os.Stderr, append([]any{ERR}, a...)...)
}

func PrintErrorf(format string, a ...any) {
	_, _ = fmt.Fprintf(os.Stderr, "%s "+format, append([]any{ERR}, a...)...)
}

func PrintInfo(a ...any) {
	_, _ = fmt.Fprintln(os.Stdout, append([]any{INFO}, a...)...)
}

func PrintInfof(format string, a ...any) {
	_, _ = fmt.Fprintf(os.Stdout, "%s "+format, append([]any{INFO}, a...)...)
}

func PrintOk(a ...any) {
	_, _ = fmt.Fprintln(os.Stdout, append([]any{OK}, a...)...)
}

func PrintOkf(format string, a ...any) {
	_, _ = fmt.Fprintf(os.Stdout, "%s "+format, append([]any{OK}, a...)...)
}

func PrintItem(a ...any) {
	_, _ = fmt.Fprintln(os.Stdout, append([]any{ITEM}, a...)...)
}

func PrintItemf(format string, a ...any) {
	_, _ = fmt.Fprintf(os.Stdout, "%s "+format, append([]any{ITEM}, a...)...)
}

func PrintReload(a ...any) {
	_, _ = fmt.Fprintln(os.Stdout, append([]any{RELOAD}, a...)...)
}

func PrintReloadf(format string, a ...any) {
	_, _ = fmt.Fprintf(os.Stdout, "%s "+format, append([]any{RELOAD}, a...)...)
}

func Black(s string) string {
	return ColorBlack + s + ColorClear
}

func Red(s string) string {
	return ColorRed + s + ColorClear
}

func Green(s string) string {
	return ColorGreen + s + ColorClear
}

func Yellow(s string) string {
	return ColorYellow + s + ColorClear
}

func Blue(s string) string {
	return ColorBlue + s + ColorClear
}

func Magenta(s string) string {
	return ColorMagenta + s + ColorClear
}

func Cyan(s string) string {
	return ColortCyan + s + ColorClear
}

func White(s string) string {
	return ColortWhite + s + ColorClear
}
