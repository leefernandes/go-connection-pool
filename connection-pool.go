package main

import (
	"errors"
	"io"
	"sync"

	"github.com/leefernandes/go-connection-pool/db"
	"golang.org/x/sync/errgroup"
)

// ConnectionPool in an unbound pool of db connections
// implemented as a linked list
type ConnectionPool interface {
	Close() error
	GetConnection() (*connection, error)
	Len() int
	ReleaseConnection(*connection)
}

// NewConnectionPool returns a ConnectionPool or error
func NewConnectionPool(addr string) (ConnectionPool, error) {
	cp := &connectionPool{
		addr: addr,
	}

	if err := cp.createConnectionAndOpen(); err != nil {
		return nil, err
	}

	return cp, nil
}

type connectionPool struct {
	addr   string
	head   *connection
	tail   *connection
	mu     sync.RWMutex
	length int
}

// Close all connections in the pool
func (cp *connectionPool) Close() error {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	// TODO confirm if goroutines are executed here
	// if called in response to SIGTERM/os.Interrupt
	eg := new(errgroup.Group)
	next := cp.head

	// TODO add a unit test for cyclical connections
	i := 0
	for next != nil && i < cp.length {
		i++
		c := next
		eg.Go(func() error {
			return cp.closeConnection(c)
		})
		next = c.next
	}

	if err := eg.Wait(); err == nil {
		return err
	}

	cp.length = 0
	cp.head = nil
	cp.tail = nil

	return nil
}

func (cp *connectionPool) Len() int {
	return cp.length
}

// GetConnection returns a connection or error
func (cp *connectionPool) GetConnection() (*connection, error) {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	// if we do not have an available connection
	// at the head of the pool create and open one
	if !cp.head.available {
		if err := cp.createConnectionAndOpen(); err != nil {
			// TODO retry on certain errors?
			return nil, err
		}
	}

	conn := cp.head
	cp.push(conn)

	return conn, nil
}

// ReleaseConnection frees the connection for reuse
func (cp *connectionPool) ReleaseConnection(conn *connection) {
	cp.unshift(conn)
}

// closeConnection closes the underlying db.Conn
func (cp *connectionPool) closeConnection(conn *connection) error {
	if err := conn.dbConn.Close(); err != nil {
		return err
	}

	conn.available = true
	conn.next = nil

	return nil
}

// createConnectionAndOpen adds a connection to the pool w/ an open *db.Conn
func (cp *connectionPool) createConnectionAndOpen() error {
	cp.length++

	dbConn := db.Conn{
		Addr: cp.addr,
	}

	if err := dbConn.Open(); err != nil {
		return err
	}

	conn := &connection{
		dbConn: dbConn,
		id:     cp.length,
	}

	if nil == cp.tail {
		cp.push(conn)
	}
	cp.unshift(conn)

	return nil
}

// push should occur when a connection is retrieved by GetConnection
func (cp *connectionPool) push(conn *connection) {
	conn.available = false
	if cp.tail == conn {
		return
	}

	if cp.head == conn && conn.next != nil {
		// move next to head
		cp.head = conn.next
	}

	// move connection to tail & unassign next
	cp.tail, conn.next = conn, nil
}

// unshift should occur when a connection is released by ReleaseConnection
func (cp *connectionPool) unshift(conn *connection) {
	conn.available = true

	if cp.head == conn {
		return
	}

	cp.head, conn.next = conn, cp.head
}

// connection linked list
type connection struct {
	available bool
	dbConn    db.Conn
	next      *connection
	id        int
}

// SendAndReceive wraps *dbConn.SendAndReceive
func (c connection) SendAndReceive(in []byte) (io.Reader, error) {
	if c.available {
		// return error when attempting to send & receive on an available (released) connection
		return nil, errors.New("released channels cannot send & receive")
	}
	return c.dbConn.SendAndReceive(in)
}
