package store

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/syndtr/goleveldb/leveldb"
	"mtlogin/internal/domain"
	"mtlogin/pkg/log"
)

// LevelDB key 白名单。token key 继续使用旧版 m-team-auth，兼容历史运行态。
const (
	// CookieDBPath 是固定的本地状态数据库目录，保存认证状态和账号绑定元数据。
	CookieDBPath = "data/cookie.db"

	tokenKey     = "m-team-auth"
	didKey       = "m-team-did"
	visitorIDKey = "m-team-visitorid"
	accountKey   = "m-team-account"
)

// LevelDBStore 使用 epoch 机制防止运行中状态清理与 in-flight Run 回写之间的竞争。
//
// Epoch 契约：
//   - SaveAuthState 仅在传入 epoch 等于当前 epoch 时写入。
//     不一致返回 saved=false，调用方丢弃本次写入。
//   - ResetAuthState 原子删除 token/DID/visitorID 并递增 epoch，
//     旧 Run 后续回写因 epoch 不匹配被拒绝。
type LevelDBStore struct {
	db    *leveldb.DB
	epoch atomic.Uint64
	mu    sync.Mutex
}

func OpenDefault() (*LevelDBStore, error) {
	return Open(CookieDBPath)
}

func Open(path string) (*LevelDBStore, error) {
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, err
	}
	if err := restrictPermissions(path); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &LevelDBStore{db: db}, nil
}

func (s *LevelDBStore) LoadAuthState() (domain.AuthState, domain.Epoch, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	state := domain.AuthState{}
	ok := false
	for key, target := range map[string]*string{
		tokenKey:     &state.Token,
		didKey:       &state.DID,
		visitorIDKey: &state.VisitorID,
	} {
		value, err := s.db.Get([]byte(key), nil)
		if errors.Is(err, leveldb.ErrNotFound) {
			continue
		}
		if err != nil {
			return domain.AuthState{}, domain.Epoch(s.epoch.Load()), false, err
		}
		*target = string(value)
		ok = true
	}
	return state, domain.Epoch(s.epoch.Load()), ok, nil
}

func (s *LevelDBStore) LoadAuthAccount() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	value, err := s.db.Get([]byte(accountKey), nil)
	if errors.Is(err, leveldb.ErrNotFound) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return string(value), nil
}

// SaveAuthState 全量覆盖 token、DID、visitorID。空字段删除对应 key。
// 仅账号密码模式调用；Token Auth 模式禁止调用此方法。
func (s *LevelDBStore) SaveAuthState(state domain.AuthState, epoch domain.Epoch) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if uint64(epoch) != s.epoch.Load() {
		return false, nil
	}
	batch := new(leveldb.Batch)
	for key, value := range map[string]string{
		tokenKey:     state.Token,
		didKey:       state.DID,
		visitorIDKey: state.VisitorID,
	} {
		if value == "" {
			batch.Delete([]byte(key))
			continue
		}
		batch.Put([]byte(key), []byte(value))
	}
	if err := s.db.Write(batch, nil); err != nil {
		return false, err
	}
	return true, nil
}

func (s *LevelDBStore) SaveAuthAccount(account string, epoch domain.Epoch) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if uint64(epoch) != s.epoch.Load() {
		return false, nil
	}
	var err error
	if account == "" {
		err = s.db.Delete([]byte(accountKey), nil)
	} else {
		err = s.db.Put([]byte(accountKey), []byte(account), nil)
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// ResetAuthState 原子删除 token/DID/visitorID/账号绑定并递增 epoch。
func (s *LevelDBStore) ResetAuthState() (domain.Epoch, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	batch := new(leveldb.Batch)
	batch.Delete([]byte(tokenKey))
	batch.Delete([]byte(didKey))
	batch.Delete([]byte(visitorIDKey))
	batch.Delete([]byte(accountKey))
	if err := s.db.Write(batch, nil); err != nil {
		return domain.Epoch(s.epoch.Load()), err
	}
	newEpoch := s.epoch.Add(1)
	log.Infof("session: identity reset new_epoch=%d", newEpoch)
	return domain.Epoch(newEpoch), nil
}

func restrictPermissions(path string) error {
	return filepath.Walk(path, func(current string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		mode := os.FileMode(0o600)
		if info.IsDir() {
			mode = 0o700
		}
		return os.Chmod(current, mode)
	})
}

func (s *LevelDBStore) Close() error {
	return s.db.Close()
}
