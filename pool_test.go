package pool

import (
	"testing"
)

type MockConnection struct {
	Id int
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
	pool := NewPool(2, new(MockFactory))
	if con, _ := pool.Allocate(); con == nil {
		t.Error(`Cannot do first allocation`)
	}
}

func TestAllocateMoreThanMax(t *testing.T) {
	pool := NewPool(2, new(MockFactory))
	pool.Allocate()
	pool.Allocate()
	if con, _ := pool.Allocate(); con == nil {
		t.Error(`Cannot do first allocation`)
	}
}

func TestRelease(t *testing.T) {
	pool := NewPool(2, new(MockFactory))
	con, _ := pool.Allocate()
	if err := pool.Release(con); err != nil {
		t.Errorf(`Cannot release just allocated connection: %s`, err)
	}
}

func TestReleaseMoreThanAllocated(t *testing.T) {
	fac := new(MockFactory)
	pool := NewPool(2, fac)
	tmp, _ := pool.Allocate()
	pool.Release(tmp)
	con, _ := fac.CreateConnection()
	if err := pool.Release(con); err == nil {
		t.Error(`Does not return error when release more than allocating`)
	}
}

func TestReleaseMoreThanMax(t *testing.T) {
	pool := NewPool(2, new(MockFactory))
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
	pool := NewPool(2, new(MockFactory))
	con, _ := pool.Allocate();
	id := con.(*MockConnection).Id
	pool.Release(con)
	if con, _ := pool.Allocate(); con.(*MockConnection).Id != id {
		t.Errorf(`trying to reallocate used connection#%d but got %d`, id, con.(*MockConnection).Id)
	}
}

func TestconnectionInterrupted(t *testing.T) {
	pool := NewPool(2, new(MockFactory))
	con, _ := pool.Allocate();
	con.(*MockConnection).Stop()
	id := con.(*MockConnection).Id
	if err := pool.Release(con); err != nil {
		t.Fatalf(`error in releasing interrupted connection: %s`, err)
	}

	tmp, _ := pool.Allocate();
	c := tmp.(*MockConnection)
	if c.Id == id {
		t.Error(`pool returned an old, interrupted connection`)
	}

	if ! c.Status {
		t.Error(`pool returned an interrupted connection`)
	}
}

func TestMaxIdle(t *testing.T) {
	pool := NewPool(2, new(MockFactory))
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
