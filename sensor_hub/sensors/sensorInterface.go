package sensors

// ISensor defines the interface that all sensor types must implement.
type ISensor interface {
	TakeReading(persist bool) error
	ToString() string
	GetName() string
	GetURL() string
}
