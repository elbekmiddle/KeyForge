package mouse

// ButtonValue mirrors keyboard.EventValue for EV_KEY button events
// (BTN_LEFT, BTN_RIGHT, BTN_EXTRA, ...).
type ButtonValue int

const (
	ButtonReleased ButtonValue = 0
	ButtonPressed  ButtonValue = 1
)

// ButtonEvent is a remappable mouse button press/release. Only these go
// through the mapping engine (doc section 14) — movement and scroll are
// passed straight through, since remapping relative motion isn't a
// "mapping" in the KEY_A -> KEY_B sense.
type ButtonEvent struct {
	Code  string // e.g. "BTN_LEFT", "BTN_EXTRA"
	Value ButtonValue
}

// MotionEvent is raw relative movement — always passed through untouched.
type MotionEvent struct {
	Axis  string // "REL_X", "REL_Y", "REL_WHEEL", "REL_HWHEEL"
	Value int32
}
