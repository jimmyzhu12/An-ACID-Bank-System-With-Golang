package clog

type Color int

const (
	Red Color = iota
	Green
	Yellow
	Blue
	Purple
	Cyan
	White
)

const (
	reset  = "\033[0m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	blue   = "\033[34m"
	purple = "\033[35m"
	cyan   = "\033[36m"
	white  = "\033[37m"
)

func getColorPreset(color Color) string {
	switch color {
	case Red:
		return red
	case Green:
		return green
	case Yellow:
		return yellow
	case Blue:
		return blue
	case Purple:
		return purple
	case Cyan:
		return cyan
	case White:
		return white
	default:
		return reset
	}
}
