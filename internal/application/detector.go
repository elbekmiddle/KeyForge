package application

type Detector interface {
	Active() (ActiveApplication, error)
}
