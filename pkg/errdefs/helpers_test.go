package errdefs

import (
	"testing"
)

func TestInvalidParameter(t *testing.T) {
	if IsInvalidParameter(errTest) {
		t.Fatalf("did not expect not found error, got %T", errTest)
	}
	e := InvalidParameter(errTest)
	if !IsInvalidParameter(e) {
		t.Fatalf("expected not found error, got: %T", e)
	}
}

func TestUnauthorized(t *testing.T) {
	if IsUnauthorized(errTest) {
		t.Fatalf("did not expect not found error, got %T", errTest)
	}
	e := Unauthorized(errTest)
	if !IsUnauthorized(e) {
		t.Fatalf("expected not found error, got: %T", e)
	}
}

func TestPayment(t *testing.T) {
	if IsPayment(errTest) {
		t.Fatalf("did not expect not found error, got %T", errTest)
	}
	e := Payment(errTest)
	if !IsPayment(e) {
		t.Fatalf("expected not found error, got: %T", e)
	}
}

func TestForbidden(t *testing.T) {
	if IsForbidden(errTest) {
		t.Fatalf("did not expect not found error, got %T", errTest)
	}
	e := Forbidden(errTest)
	if !IsForbidden(e) {
		t.Fatalf("expected not found error, got: %T", e)
	}
}

func TestNotFound(t *testing.T) {
	if IsNotFound(errTest) {
		t.Fatalf("did not expect not found error, got %T", errTest)
	}
	e := NotFound(errTest)
	if !IsNotFound(e) {
		t.Fatalf("expected not found error, got: %T", e)
	}
}

func TestUnprocessableEntity(t *testing.T) {
	if IsUnprocessableEntity(errTest) {
		t.Fatalf("did not expect not found error, got %T", errTest)
	}
	e := UnprocessableEntity(errTest)
	if !IsUnprocessableEntity(e) {
		t.Fatalf("expected not found error, got: %T", e)
	}
}

func TestDataBase(t *testing.T) {
	if IsDataBase(errTest) {
		t.Fatalf("did not expect not found error, got %T", errTest)
	}
	e := DataBase(errTest)
	if !IsDataBase(e) {
		t.Fatalf("expected not found error, got: %T", e)
	}
}

func TestRedis(t *testing.T) {
	if IsRedis(errTest) {
		t.Fatalf("did not expect not found error, got %T", errTest)
	}
	e := Redis(errTest)
	if !IsRedis(e) {
		t.Fatalf("expected not found error, got: %T", e)
	}
}
