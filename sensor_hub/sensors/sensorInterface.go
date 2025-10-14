package sensors

type ISensor interface {
	TakeReading(persist bool) error
	ToString() string
	GetName() string
	GetURL() string
}
