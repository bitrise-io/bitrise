package log

// Logger ...
type Logger interface {
	Print(f Formatable)
}

// Formatable ...
type Formatable interface {
	String() string
	JSON() string
}
