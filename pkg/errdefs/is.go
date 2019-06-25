package errdefs

func IsInvalidParameter(err error) bool {
	_, ok := err.(ErrInvalidParameter)
	return ok
}

func IsUnauthorized(err error) bool {
	_, ok := err.(ErrUnauthorized)
	return ok
}

func IsPayment(err error) bool {
	_, ok := err.(ErrPayment)
	return ok
}

func IsForbidden(err error) bool {
	_, ok := err.(ErrForbidden)
	return ok
}

func IsNotFound(err error) bool {
	_, ok := err.(ErrNotFound)
	return ok
}

func IsUnprocessableEntity(err error) bool {
	_, ok := err.(ErrUnprocessableEntity)
	return ok
}

func IsDataBase(err error) bool {
	_, ok := err.(ErrDataBase)
	return ok
}

func IsRedis(err error) bool {
	_, ok := err.(ErrRedis)
	return ok
}
