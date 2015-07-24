# Simple implementation of connection pool

## Synopsis

```go
package main

import (
        "net/http"
        "github.com/Ronmi/pool"
        "some.where/myproj/mypkg"
)

type MyConnectionFactory struct {}

func (f *MyConnectionFactory) CreateConnection() (interface{}, error) {
        conn, err := mypkg.Open(args...)
        return conn, err
}

func (f *MyConnectionFactory) CloseConnection(conn interface{}) error {
        err := conn.(*mypkg.MyConnection).Close()
        return err
}

func (f *MyConnectionFactory) CheckConnection(conn interface{}) bool {
        // mypkg.MyConnection does not provide fast method to check connection usability.
        return true
}


type myprog struct {
	p pool.Pool
}

func (m *myprog) Do(w http.ResponseWriter, r *http.Request) {
	con, err := m.p.Allocate()
	if err != nil {
		// do error handling stuff
	}
	defer m.p.Release(con)

	// do real stuff
}

func main() {
	p := pool.NewPool(3, new(MyConnectionFactory))
	m := &myprog{p}
	http.Handle("/", http.HandlerFunc(m.Do))
	http.ListenAndServe(":80", nil)
}
```


## License
Any version of MIT, GPL or LGPL
