## Connection Pool implemented as a Linked List

```go
go test -v .
```

## Example Usage
```go
pool, err := NewConnectionPool("localhost:27017")
if err != nil {
	log.Fatal(err)
}

// create & return a new connection
conn1, err := pool.GetConnection()
if err != nil {
	log.Fatal(err)
}

// create & return a second connection
conn2, err := pool.GetConnection()
if err != nil {
	log.Fatal(err)
}

// release the first connection
pool.ReleaseConnection(conn1)

// returns the first connection
conn3, err := pool.GetConnection()
if err != nil {
	log.Fatal(err)
}

// close all connections
if err := pool.Close(); err != nil {
	log.Fatal(err)
}

```