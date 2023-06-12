package platform

type BUTTON_STATE uint8
const (
	BS_UP BUTTON_STATE = iota
	BS_PRESSED
	BS_DOWN
	BS_RELEASED
)

func ButtonMapUpdate[K comparable](m map[K]BUTTON_STATE) {
	for btn, state := range m {
		if state == BS_PRESSED {
			m[btn] = BS_DOWN
		} else if state == BS_RELEASED {
			m[btn] = BS_UP
		}
	}
}

func KeyboardPressed(key uint32) bool {
	state, ok := Platform.Keyboard[key]
	if !ok { return false }
	return state == BS_PRESSED
}

func KeyboardReleased(key uint32) bool {
	state, ok := Platform.Keyboard[key]
	if !ok { return false }
	return state == BS_RELEASED
}

func MousePressed(btn uint8) bool {
	state, ok := Platform.Mouse[btn]
	if !ok { return false }
	return state == BS_PRESSED
}

func MouseReleased(btn uint8) bool {
	state, ok := Platform.Mouse[btn]
	if !ok { return false }
	return state == BS_RELEASED
}

func WindowWidth() float32 {
	w, _ := Platform.Window.GetSize()
	return float32(w)
}

func WindowHeight() float32 {
	_, h := Platform.Window.GetSize()
	return float32(h)
}

func TextWidth(text string) float32 {
	return TextMetrics(text).X
}

func TextHeight(text string) float32 {
	return TextMetrics(text).Y
}
