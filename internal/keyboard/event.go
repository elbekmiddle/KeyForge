package keyboard

type EventValue int

const (
	KeyReleased EventValue = 0
	KeyPressed  EventValue = 1
	KeyRepeated EventValue = 2
)

type KeyEvent struct {
	Code  string
	Value EventValue
}
