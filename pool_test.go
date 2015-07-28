package pool

import (
	"testing"
)

type MockConnection struct {
	Id     int
	Status bool
}

func (c *MockConnection) Stop() {
	c.Status = false
}

type MockFactory struct {
	counter int
}

func (f *MockFactory) CreateConnection() (interface{}, error) {
	f.counter++
	return &MockConnection{Id: f.counter, Status: true}, nil
}

func (f *MockFactory) CloseConnection(con interface{}) error {
	con.(*MockConnection).Stop()
	return nil
}

func (f *MockFactory) CheckConnection(con interface{}) bool {
	return con.(*MockConnection).Status
}

func TestAllocateOne(t *testing.T) {
	pool := New(2, 3, new(MockFactory))
	if con, _ := pool.Allocate(); con == nil {
		t.Error(`Cannot do first allocation`)
	}
}

func TestAllocateMoreThanMax(t *testing.T) {
	pool := New(2, 3, new(MockFactory))
	pool.Allocate()
	pool.Allocate()
	if con, _ := pool.Allocate(); con == nil {
		t.Error(`Cannot do first allocation`)
	}
}

func TestRelease(t *testing.T) {
	pool := New(2, 3, new(MockFactory))
	con, _ := pool.Allocate()
	if err := pool.Release(con); err != nil {
		t.Errorf(`Cannot release just allocated connection: %s`, err)
	}
}

func TestReleaseMoreThanAllocated(t *testing.T) {
	fac := new(MockFactory)
	pool := New(2, 3, fac)
	tmp, _ := pool.Allocate()
	pool.Release(tmp)
	con, _ := fac.CreateConnection()
	if err := pool.Release(con); err == nil {
		t.Error(`Does not return error when release more than allocating`)
	}
}

func TestReleaseMoreThanMax(t *testing.T) {
	pool := New(2, 3, new(MockFactory))
	var f = func() *MockConnection {
		tmp, _ := pool.Allocate()
		return tmp.(*MockConnection)
	}
	cons := [3]*MockConnection{
		f(),
		f(),
		f(),
	}

	for _, v := range cons {
		if err := pool.Release(v); err != nil {
			t.Errorf(`Failed to release connection#%d: %s`, v.Id, err)
		}
	}
}

func TestReallocateQueuedconnection(t *testing.T) {
	pool := New(2, 3, new(MockFactory))
	con, _ := pool.Allocate()
	id := con.(*MockConnection).Id
	pool.Release(con)
	if con, _ := pool.Allocate(); con.(*MockConnection).Id != id {
		t.Errorf(`trying to reallocate used connection#%d but got %d`, id, con.(*MockConnection).Id)
	}
}

func TestconnectionInterrupted(t *testing.T) {
	pool := New(2, 3, new(MockFactory))
	con, _ := pool.Allocate()
	con.(*MockConnection).Stop()
	id := con.(*MockConnection).Id
	if err := pool.Release(con); err != nil {
		t.Fatalf(`error in releasing interrupted connection: %s`, err)
	}

	tmp, _ := pool.Allocate()
	c := tmp.(*MockConnection)
	if c.Id == id {
		t.Error(`pool returned an old, interrupted connection`)
	}

	if !c.Status {
		t.Error(`pool returned an interrupted connection`)
	}
}

func TestMaxIdle(t *testing.T) {
	pool := New(2, 30, new(MockFactory))
	var f = func() *MockConnection {
		tmp, _ := pool.Allocate()
		return tmp.(*MockConnection)
	}
	cons := []*MockConnection{
		f(),
		f(),
		f(),
		f(),
		f(),
		f(),
	}

	counter := 0
	for _, v := range cons {
		if pool.Release(v); v.Status {
			counter++
			if counter > 2 {
				t.Errorf(`Exceeding max idle when releasing connection#%d`, v.Id)
			}
		}
	}

	if counter != 2 {
		t.Errorf(`expected idle connections is 2, got %d`, counter)
	}
}

func TestMaxRun(t *testing.T) {
	pool := New(2, 2, new(MockFactory))
	ch := make(chan *MockConnection, 3)

	tmp, _ := pool.Allocate()
	ch <- tmp.(*MockConnection)
	tmp, _ = pool.Allocate()
	ch <- tmp.(*MockConnection)

	go func() {
		tmp, _ := pool.Allocate()
		ch <- tmp.(*MockConnection)
	}()

	for range [3]int{1, 2, 3} {
		c := <-ch
		pool.Release(c)
		if c.Id != 1 && c.Id != 2 {
			t.Errorf("Expected connecion id is 1 or 2, got %d", c.Id)
		}
	}
}
