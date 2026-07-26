package store

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
	"mtlogin/internal/domain"
)

func TestStoreEpochPreventsStaleWrites(t *testing.T) {
	store := openTestStore(t)
	state := domain.AuthState{Token: "token1", DID: "did1", VisitorID: "visitor1"}

	_, epoch, _, err := store.LoadAuthState()
	if err != nil {
		t.Fatal(err)
	}
	if saved, err := store.SaveAuthState(state, epoch); err != nil || !saved {
		t.Fatalf("save auth state: saved=%v err=%v", saved, err)
	}
	if _, err := store.ResetAuthState(); err != nil {
		t.Fatal(err)
	}
	if saved, err := store.SaveAuthState(domain.AuthState{Token: "stale"}, epoch); err != nil || saved {
		t.Fatalf("stale write should be ignored: saved=%v err=%v", saved, err)
	}
	got, _, ok, err := store.LoadAuthState()
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatalf("expected reset state after stale write, got %+v", got)
	}
}

func TestStoreReadsAuthKey(t *testing.T) {
	dir := t.TempDir()
	db, err := leveldb.OpenFile(filepath.Join(dir, "db"), nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Put([]byte("m-team-auth"), []byte("token"), nil); err != nil {
		t.Fatal(err)
	}
	if err := db.Put([]byte("m-team-did"), []byte("did"), nil); err != nil {
		t.Fatal(err)
	}
	if err := db.Put([]byte("m-team-visitorid"), []byte("visitor"), nil); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	store, err := Open(filepath.Join(dir, "db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	state, _, ok, err := store.LoadAuthState()
	if err != nil {
		t.Fatal(err)
	}
	if !ok || state.Token != "token" || state.DID != "did" || state.VisitorID != "visitor" {
		t.Fatalf("auth key must be loaded, ok=%v state=%+v", ok, state)
	}
}

func TestStorePersistsAndResetsAuthAccount(t *testing.T) {
	store := openTestStore(t)
	_, epoch, _, err := store.LoadAuthState()
	if err != nil {
		t.Fatal(err)
	}
	if saved, err := store.SaveAuthAccount("alice", epoch); err != nil || !saved {
		t.Fatalf("save account: saved=%v err=%v", saved, err)
	}
	if account, err := store.LoadAuthAccount(); err != nil || account != "alice" {
		t.Fatalf("load account: account=%q err=%v", account, err)
	}
	newEpoch, err := store.ResetAuthState()
	if err != nil {
		t.Fatal(err)
	}
	if account, err := store.LoadAuthAccount(); err != nil || account != "" {
		t.Fatalf("reset must clear account: account=%q err=%v", account, err)
	}
	if saved, err := store.SaveAuthAccount("bob", epoch); err != nil || saved {
		t.Fatalf("stale account write should be ignored: saved=%v err=%v", saved, err)
	}
	if saved, err := store.SaveAuthAccount("bob", newEpoch); err != nil || !saved {
		t.Fatalf("current account write failed: saved=%v err=%v", saved, err)
	}
}

func TestStoreConcurrentResetDoesNotAllowStaleWrite(t *testing.T) {
	for i := 0; i < 100; i++ {
		store := openTestStore(t)
		_, epoch, _, err := store.LoadAuthState()
		if err != nil {
			t.Fatal(err)
		}

		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			time.Sleep(time.Microsecond)
			_, _ = store.SaveAuthState(domain.AuthState{Token: "stale", DID: "stale-did", VisitorID: "stale-visitor"}, epoch)
		}()
		go func() {
			defer wg.Done()
			_, _ = store.ResetAuthState()
		}()
		wg.Wait()

		got, currentEpoch, ok, err := store.LoadAuthState()
		if err != nil {
			t.Fatal(err)
		}
		if currentEpoch > epoch && ok {
			t.Fatalf("stale write survived reset: epoch=%d current=%d state=%+v", epoch, currentEpoch, got)
		}
		_ = store.Close()
	}
}

func TestStoreUsesPrivateFilesystemPermissions(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "db")
	store, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o700 {
		t.Fatalf("database directory permissions = %04o, want 0700", got)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		fileInfo, err := entry.Info()
		if err != nil {
			t.Fatal(err)
		}
		if got := fileInfo.Mode().Perm(); got != 0o600 {
			t.Fatalf("database file %s permissions = %04o, want 0600", entry.Name(), got)
		}
	}
}

func openTestStore(t *testing.T) *LevelDBStore {
	t.Helper()
	store, err := Open(filepath.Join(t.TempDir(), "db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return store
}
