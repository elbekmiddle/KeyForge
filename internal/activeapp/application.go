package activeapp

type Application struct {
	Name       string
	Class      string
	Executable string
}

func (a Application) IsZero() bool {
	return a.Name == "" &&
		a.Class == "" &&
		a.Executable == ""
}
