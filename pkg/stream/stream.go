package stream

type StreamBp[T any] interface {
}

type streamBp[T any] struct {
	name string
	lbp  StreamBp[any]
	mf   MapFunction[any, T]
}

// Step1: Run bulid stream function to collect blueprint

type MapFunction[T, R any] func(t T) R

func Map[T, R any](srcStream StreamBp[T], mapFunction MapFunction[T, R]) (destStream *streamBp[R]) {
	mf := func(a any) R {
		t := a.(T)
		return mapFunction(t)
	}

	destStream = &streamBp[R]{
		lbp: srcStream,
		mf:  mf,
	}
	return
}

func (s *streamBp[T]) SetName(name string) {
	s.name = name
}

var blueprints streamBp[any]

func Apply() {
	// assemble each blueprint to be a fission function

}
