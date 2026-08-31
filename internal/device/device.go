package device

type Type string

const (
	TypeKeyboard Type = "keyboard"
	TypeMouse    Type = "mouse"
	TypeUnknown  Type = "unknown"
)

type Device struct {
	Name         string
	Type         Type
	Path         string
	VendorID     string
	ProductID    string
	Bus          string
	Manufacturer string
}
