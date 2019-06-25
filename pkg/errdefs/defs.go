package errdefs

// 400
type ErrInvalidParameter interface {
	InvalidParameter()
}

// 401
type ErrUnauthorized interface {
	Unauthorized()
}

// 402
type ErrPayment interface {
	Payment()
}

// 403
type ErrForbidden interface {
	Forbidden()
}

// 404
type ErrNotFound interface {
	NotFound()
}

// 422
type ErrUnprocessableEntity interface {
	UnprocessableEntity()
}

// database出錯
type ErrDataBase interface {
	DataBase()
}

// redis出錯
type ErrRedis interface {
	Redis()
}
