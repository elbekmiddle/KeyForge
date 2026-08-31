package activeapp

type Application struct {
	Name       string
	Class      string
	Executable string
	PID        int
}

type Detector interface {
	Current() (Application, error)
}
