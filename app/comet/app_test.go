package comet

import (
	"fmt"
	"os"
	"testing"
	"time"
)

/*
testing template for other case
NOTE: make sure, your util have been testing successfully as well
*/

func TestMain(m *testing.M) {
	fmt.Println("TestMain start setup")
	exitCode := m.Run()
	fmt.Println("TestMain end")
	os.Exit(exitCode)
}

func TestTask1(t *testing.T) {
	resource := true

	defer func() {
		resource = false
	}()
	t.Run("subtask-1", func(t *testing.T) {
		t.Parallel() //push subtest to queue
	})
	t.Run("subtask-2", func(t *testing.T) {
		t.Parallel() //push subtest to queue
	})
}

func TestTask2(t *testing.T) {
	//run t.Parallel()
	t.Log("TestTask2 preparation")
	resource := true //

	defer func() {
		resource = false
	}()

	p1 := func(t *testing.T) {
		t.Parallel()
	}

	t.Run("gorup all subtasks under same task", func(t *testing.T) {
		t.Run("SubProcess-1", p1)
		t.Run("SubProcess-2", func(t *testing.T) {
			t.Parallel()
		})
	})
}

func onWork(sec int, taskName string, symbol rune) {

	done := time.After(time.Second * time.Duration(sec))
	for {
		time.Sleep(time.Second * 1)
		select {
		case <-done:
			return //
		default:
			fmt.Print(symbol)
		}
	}
	fmt.Printf("\n")
}

/*
func TestTask2(t * testing.T){

}

func TestParallTask1( t *testing.T){
	t.Parallel()
}
*/

func TestClosureError(t *testing.T) {
	for i := 0; i < 10; i++ {
		t.Run(fmt.Sprintf("i=%d", i), func(t *testing.T) {
			t.Parallel()
			/*
				can pass fun goes here ..
			*/
		})
	}
}

/*
type sqlDB interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

type mockDB struct{}

func (mdb *mockDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return nil, nil
}

func saveUser(db sqlDB, user *User) error {
	if user.EmailAddress == "" {
		return errors.New("user requires an email")
	}
	if len(user.Password) < 8 {
		return errors.New("user password requires at least 8 characters")
	}
	hashedPassword, err := hash(user.Password)
	if err != nil {
		return err
	}
		_, err := db.Exec(` INSERT INTO member (password, email, createat) VALUES ($1, $2, $3);`, hashedPassword, user.EmailAddress, time.Now())
	return err
}

func hash(password string) (string, error) {
	const cost = 10
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	return string(bytes), err
}
*/
