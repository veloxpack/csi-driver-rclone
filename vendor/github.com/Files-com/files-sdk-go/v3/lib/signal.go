package lib

type Signal struct {
	C chan bool
}

func (s *Signal) Init() *Signal {
	s.C = make(chan bool)
	return s
}

func (s *Signal) Called() bool {
	select {
	case <-s.C:
		return true
	default:
		return false
	}
}

func (s *Signal) Call() {
	select {
	case <-s.C:
	default:
		close(s.C)
	}
}

func (s *Signal) Clear() {
	s.Init()
}
