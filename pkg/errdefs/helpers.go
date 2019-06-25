package errdefs

type causer interface {
	Code() int
	Cause() error
}

type errInvalidParameter struct {
	error
	code int
}

func (errInvalidParameter) InvalidParameter() {}
func (e errInvalidParameter) Code() int       { return e.code }
func (e errInvalidParameter) Cause() error    { return e.error }

func InvalidParameter(err error, c ...int) error {
	var code int
	if err == nil || IsInvalidParameter(err) {
		return err
	}
	if c != nil {
		code = c[0]
	}
	return errInvalidParameter{error: err, code: code}
}

type errUnauthorized struct {
	error
	code int
}

func (errUnauthorized) Unauthorized()  {}
func (e errUnauthorized) Code() int    { return e.code }
func (e errUnauthorized) Cause() error { return e.error }

func Unauthorized(err error, c ...int) error {
	var code int
	if err == nil || IsUnauthorized(err) {
		return err
	}
	if c != nil {
		code = c[0]
	}
	return errUnauthorized{error: err, code: code}
}

type errPayment struct {
	error
	code int
}

func (errPayment) Payment()       {}
func (e errPayment) Code() int    { return e.code }
func (e errPayment) Cause() error { return e.error }

func Payment(err error, c ...int) error {
	var code int
	if err == nil || IsPayment(err) {
		return err
	}
	if c != nil {
		code = c[0]
	}
	return errPayment{error: err, code: code}
}

type errForbidden struct {
	error
	code int
}

func (errForbidden) Forbidden()     {}
func (e errForbidden) Code() int    { return e.code }
func (e errForbidden) Cause() error { return e.error }

func Forbidden(err error, c ...int) error {
	var code int
	if err == nil || IsForbidden(err) {
		return err
	}
	if c != nil {
		code = c[0]
	}
	return errForbidden{error: err, code: code}
}

type errNotFound struct {
	error
	code int
}

func (errNotFound) NotFound()      {}
func (e errNotFound) Code() int    { return e.code }
func (e errNotFound) Cause() error { return e.error }

func NotFound(err error, c ...int) error {
	var code int
	if err == nil || IsNotFound(err) {
		return err
	}
	if c != nil {
		code = c[0]
	}
	return errNotFound{error: err, code: code}
}

type errUnprocessableEntity struct {
	error
	code int
}

func (errUnprocessableEntity) UnprocessableEntity() {}
func (e errUnprocessableEntity) Code() int          { return e.code }
func (e errUnprocessableEntity) Cause() error       { return e.error }

func UnprocessableEntity(err error, c ...int) error {
	var code int
	if err == nil || IsUnprocessableEntity(err) {
		return err
	}
	if c != nil {
		code = c[0]
	}
	return errUnprocessableEntity{error: err, code: code}
}

type errDataBase struct {
	error
	code int
}

func (errDataBase) DataBase()      {}
func (e errDataBase) Code() int    { return e.code }
func (e errDataBase) Cause() error { return e.error }

func DataBase(err error, c ...int) error {
	var code int
	if err == nil || IsDataBase(err) {
		return err
	}
	if c != nil {
		code = c[0]
	}
	return errDataBase{error: err, code: code}
}

type errRedis struct {
	error
	code int
}

func (errRedis) Redis()         {}
func (e errRedis) Code() int    { return e.code }
func (e errRedis) Cause() error { return e.error }

func Redis(err error, c ...int) error {
	var code int
	if err == nil || IsRedis(err) {
		return err
	}
	if c != nil {
		code = c[0]
	}
	return errRedis{error: err, code: code}
}
