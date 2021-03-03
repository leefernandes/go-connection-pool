package main

import (
	"io/ioutil"
	"testing"
)

// TODO concurrency test

func TestCreatePool(t *testing.T) {
	t.Run("CreatePool", func(t *testing.T) {
		_, err := NewConnectionPool("localhost")
		if err != nil {
			t.Errorf("🛑 test %s expected nil err, got %w", t.Name(), err)
		}
	})
}

func TestSendAndReceiveOneConnection(t *testing.T) {
	t.Run("SendAndReceiveOneConnection", func(t *testing.T) {
		pool, err := NewConnectionPool("localhost")
		if err != nil {
			panic(err)
		}

		conn, err := pool.GetConnection()
		if err != nil {
			t.Errorf("🛑 test %s expected nil err, got %w", t.Name(), err)
		}

		if res, err := conn.SendAndReceive([]byte("hello")); err != nil {
			t.Errorf("🛑 test %s expected nil err, got %w", t.Name(), err)
		} else {
			b, err := ioutil.ReadAll(res)
			if err != nil {
				t.Errorf("🛑 test %s expected nil err, got %w", t.Name(), err)
			}

			if string(b) != "hello👋" {
				t.Errorf("🛑 test %s expectation fail, got %s", t.Name(), string(b))
			}
		}

		if err := pool.Close(); err != nil {
			t.Errorf("🛑 test %s expected nil err, got %w", t.Name(), err)
		}
	})
}

func TestConnectionReuse(t *testing.T) {
	t.Run("ConnectionReuse", func(t *testing.T) {
		pool, err := NewConnectionPool("localhost")
		if err != nil {
			panic(err)
		}

		conn1, err := pool.GetConnection()
		if err != nil {
			t.Errorf("🛑 test %s expected nil err, got %w", t.Name(), err)
		}

		if conn1.id != 1 {
			t.Errorf("🛑 test %s expected id of %d got %d", t.Name(), 1, conn1.id)
		}

		conn2, err := pool.GetConnection()
		if err != nil {
			t.Errorf("🛑 test %s expected nil err, got %w", t.Name(), err)
		}

		if conn2.id != 2 {
			t.Errorf("🛑 test %s expected id of %d got %d", t.Name(), 2, conn2.id)
		}

		conn3, err := pool.GetConnection()
		if err != nil {
			t.Errorf("🛑 test %s expected nil err, got %w", t.Name(), err)
		}

		if conn3.id != 3 {
			t.Errorf("🛑 test %s expected id of %d got %d", t.Name(), 3, conn3.id)
		}

		pool.ReleaseConnection(conn2)

		conn4, err := pool.GetConnection()
		if err != nil {
			t.Errorf("🛑 test %s expected nil err, got %w", t.Name(), err)
		}

		if conn4.id != 2 {
			t.Errorf("🛑 test %s expected to reuse %d got %d", t.Name(), 2, conn4.id)
		}

		pool.ReleaseConnection(conn1)

		conn5, err := pool.GetConnection()
		if err != nil {
			t.Errorf("🛑 test %s expected nil err, got %w", t.Name(), err)
		}

		if conn5.id != 1 {
			t.Errorf("🛑 test %s expected if of %d got %d", t.Name(), 1, conn5.id)
		}

		if err := pool.Close(); err != nil {
			t.Errorf("🛑 test %s expected nil err, got %w", t.Name(), err)
		}
	})
}
