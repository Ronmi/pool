package pool

import (
	"net/http"
)

// Fake connection object
type MyConnection struct {
	Closed bool
}

func (c *MyConnection) Close() error {
	c.Closed = true
	return nil
}

// Basic implementation of ConnectionFactory
type MyConnectionFactory struct{}

func (f *MyConnectionFactory) CreateConnection() (interface{}, error) {
	conn := new(MyConnection)
	return conn, nil
}

func (f *MyConnectionFactory) CloseConnection(conn interface{}) error {
	err := conn.(*MyConnection).Close()
	return err
}

func (f *MyConnectionFactory) CheckConnection(conn interface{}) bool {
	return !conn.(*MyConnection).Closed
}

// My web handler
type myprog struct {
	p Pool
}

func (m *myprog) Do(w http.ResponseWriter, r *http.Request) {
	con, err := m.p.Allocate()
	if err != nil {
		// do error handling stuff
	}
	defer m.p.Release(con)

	// do real stuff
}

func Example() {
	p := New(3, 10, new(MyConnectionFactory))
	m := &myprog{p}
	http.Handle("/", http.HandlerFunc(m.Do))
	http.ListenAndServe(":80", nil)
}
