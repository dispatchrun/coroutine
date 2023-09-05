package coroutine

func Run[R, S any](g Generator[R, S], f func(R) S) {
	for g.Next() {
		r := g.Recv()
		s := f(r)
		g.Send(s)
	}
}
