package face

// Displayer is the interface for rendering pixels to a display.
type Displayer interface {
	SetPixel(x, y int, value byte)
	Clear()
	Show() error
}
