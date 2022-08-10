package ext

// Service represents generic interface to an embedded service running within anycable-go
type Service interface {
	Start() error
	WaitReady() error
	Shutdown() error
}
