package trino_cluster

import (
	"testing"
	"time"
)

func TestLockOperationsSerializesSameCluster(t *testing.T) {
	unlock := LockOperations("cluster-a")

	acquired := make(chan struct{})
	go func() {
		LockOperations("cluster-a")()
		close(acquired)
	}()

	select {
	case <-acquired:
		t.Fatal("second lock on the same cluster was acquired while the first one is held")
	case <-time.After(100 * time.Millisecond):
	}

	unlock()

	select {
	case <-acquired:
	case <-time.After(5 * time.Second):
		t.Fatal("second lock on the same cluster was not acquired after unlock")
	}
}

func TestLockOperationsIndependentClusters(t *testing.T) {
	unlock := LockOperations("cluster-b")
	defer unlock()

	acquired := make(chan struct{})
	go func() {
		LockOperations("cluster-c")()
		close(acquired)
	}()

	select {
	case <-acquired:
	case <-time.After(5 * time.Second):
		t.Fatal("lock on another cluster was blocked")
	}
}
