/*
Pool is a factory-based connnection pool implementation.

Synopsis

This implementation is thread-safe by using channel to ensure only one goroutine
can do the Allocate() or Release() work.
*/
package pool

import (
	"container/list"
)

// ErrRelease is error message telling error when releasing connection.
type ErrRelease string

func (e ErrRelease) Error() string {
	return `Cannot release connection: ` + string(e)
}

// type Pool defines what action you can do with connection pool.
type Pool interface {
	// Allocate() is used to allocating a connection.
	Allocate() (interface{}, error)

	// Release() is used to release a connection.
	Release(interface{}) error
}

/*
ConnectionFactory defines basic factory facility
*/
type ConnectionFactory interface {
	// CreateConnection() a new connection
	CreateConnection() (interface{}, error)

	// CloseConnection() an unused connection
	CloseConnection(interface{}) error

	// CheckConnection() is used to check connection usability when allocating idle connection.
	//
	// This method is called in lock state. Doing time-consuming task in it might cause
	// unwanted behavier.
	CheckConnection(interface{}) bool
}

type pool struct {
	lock      chan int
	idle      *list.List
	len       int
	max       int
	allocated chan int
	factory   ConnectionFactory
}

// New() is used to create a new connection pool.
func New(max_idle, max_running int, factory ConnectionFactory) Pool {
	return &pool{
		lock:      make(chan int, 1),
		idle:      list.New(),
		max:       max_idle,
		len:       0,
		allocated: make(chan int, max_running),
		factory:   factory,
	}
}

func (p *pool) old() (ret interface{}) {
	if p.len > 0 {
		tmp := p.idle.Front()
		p.idle.Remove(tmp)
		p.len--
		if p.factory.CheckConnection(tmp.Value) {
			ret = tmp.Value
		}
	}
	return
}

func (p *pool) new() (ret interface{}, err error) {
	ret, err = p.factory.CreateConnection()
	return
}

func (p *pool) Allocate() (ret interface{}, err error) {
	p.allocated <- 1
	p.lock <- 1
	defer func() { <-p.lock }()

	if p.len > 0 {
		for ret == nil && p.len > 0 {
			ret = p.old()
		}

		if ret == nil {
			ret, err = p.new()
		}
		return
	}

	ret, err = p.new()
	return
}

func (p *pool) Release(con interface{}) (err error) {
	p.lock <- 1
	defer func() { <-p.lock }()

	if len(p.allocated) < 1 {
		err = ErrRelease(`Yet allocated anything!`)
		return
	}

	defer func() { <-p.allocated }()
	if p.len < p.max {
		p.idle.PushBack(con)
		p.len++
		return
	}

	err = p.factory.CloseConnection(con)
	return
}
