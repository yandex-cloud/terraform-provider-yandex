package pluginsdk

type Wrapper[V any, T any] interface {
	GetValue() V
	*T
}

func getOrDefault[V any, T any, PT Wrapper[V, T]](wrapper PT, def V) V {
	if wrapper == nil {
		return def
	}
	return wrapper.GetValue()
}
