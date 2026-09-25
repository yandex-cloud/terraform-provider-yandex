package value

func CastTo(v Value, dst interface{}) error {
	if dst == nil {
		return errNilDestination
	}
	if ptr, has := dst.(*Value); has {
		*ptr = v

		return nil
	}

	return v.castTo(dst)
}
