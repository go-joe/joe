package joe

type Adapter interface {
	NextMessage() <-chan string
	Send(msg string) error
}
