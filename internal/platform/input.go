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

func (p *platform) KeyboardPressed(key uint32) bool {
	state, ok := p.Keyboard[key]
	if !ok { return false }
	return state == BS_PRESSED
}

func (p *platform) KeyboardReleased(key uint32) bool {
	state, ok := p.Keyboard[key]
	if !ok { return false }
	return state == BS_RELEASED
}

func (p *platform) MousePressed(btn uint8) bool {
	state, ok := p.Mouse[btn]
	if !ok { return false }
	return state == BS_PRESSED
}

func (p *platform) MouseReleased(btn uint8) bool {
	state, ok := p.Mouse[btn]
	if !ok { return false }
	return state == BS_RELEASED
}

func (p *platform) MouseDown(btn uint8) bool {
	state, ok := p.Mouse[btn]
	if !ok { return false }
	return state == BS_DOWN
}

func (p *platform) MouseUp(btn uint8) bool {
	state, ok := p.Mouse[btn]
	if !ok { return false }
	return state == BS_UP
}

