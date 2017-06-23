package login

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"sync"
	"time"

	"github.com/satori/go.uuid"

	"bulkytree.com/sevenhearts/auth"
)

const maxLifetime int64 = 5
const defaultLifetime int64 = 5

type UUID uuid.UUID

func (u UUID) String() string {
	return uuid.UUID(u).String()
}

var sessions map[UUID]session
var SessionManager *sessionManager

func init() {
	sessions = NewSessions()
	SessionManager = NewSessionManager(defaultLifetime, sessions)
	SessionManager.SessionGC()
}

type sessionManager struct {
	lock       sync.RWMutex
	sessions   map[UUID]session
	gcLifetime int64
}

type session interface {
	Set(key string, value interface{})  //set session value
	Get(key string) (interface{}, bool) //get session value
	Delete(key string)                  //delete session value

	SessionKey() UUID //back current sessionID
	//setSessionKey(skey UUID)
	Expiry() int64 //back current expiry
	//setExpiry(v int64)
}

type sessionObj struct {
	session    map[string]interface{}
	sessionKey UUID
	expiry     int64
}

func (s sessionObj) Get(key string) (interface{}, bool) {
	v, ok := s.session[key]
	return v, ok
}

func (s sessionObj) Set(key string, value interface{}) {
	s.session[key] = value
}

func (s sessionObj) Delete(key string) {
	if _, ok := s.session[key]; ok {
		delete(s.session, key)
	}
}

func (s sessionObj) SessionKey() UUID {
	return s.sessionKey
}

func (s sessionObj) Expiry() int64 {
	return s.expiry
}

func (s *sessionObj) setSessionKey(sKey UUID) {
	s.sessionKey = sKey
}

func (s *sessionObj) setExpiry(v int64) {
	s.expiry = v
}

func (m *sessionManager) NewSessionKey() UUID {
	return UUID(uuid.NewV4())
}

func (m *sessionManager) sessionInit(sKey UUID) session {
	s := sessionObj{}
	s.session = make(map[string]interface{})
	s.setSessionKey(sKey)
	s.setExpiry(time.Now().Unix() + m.gcLifetime)
	m.sessions[sKey] = s
	return s
}

func (m *sessionManager) SessionRead(sKey UUID) session {
	m.lock.RLock()
	s, ok := m.sessions[sKey]
	m.lock.RUnlock()

	if !ok {
		m.lock.Lock()
		defer m.lock.Unlock()
		return m.sessionInit(sKey)
	}
	return s
}

func (m *sessionManager) SessionDestroy(sKey UUID) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if _, ok := m.sessions[sKey]; ok {
		delete(m.sessions, sKey)
	}
}

func (m *sessionManager) SessionGC() {
	if m.gcLifetime == 0 {
		return
	}

	for key := range m.sessions {
		if m.sessions[key].Expiry() <= time.Now().Unix() {
			m.SessionDestroy(key)
		}
	}

	fmt.Printf("\nIn GC!, %v\n timenow: %v\n", m.sessions, time.Now().Unix())

	time.AfterFunc(time.Duration(m.gcLifetime)*time.Second, func() { m.SessionGC() })
}

func (m *sessionManager) SessionStart() (session session) {
	m.lock.Lock()
	defer m.lock.Unlock()

	sKey := m.NewSessionKey()
	session = m.sessionInit(sKey)
	fmt.Printf("session: %v, sessionKey: %v, sKey: %v", session, session.SessionKey(), sKey)
	return session
}

func NewSessionManager(gcLifetime int64, sessions map[UUID]session) *sessionManager {
	if maxLifetime == 0 {
		return &sessionManager{gcLifetime: maxLifetime, sessions: sessions}
	}

	if gcLifetime > maxLifetime {
		gcLifetime = maxLifetime
	}
	return &sessionManager{gcLifetime: gcLifetime, sessions: sessions}
}

func NewSessions() map[UUID]session {
	return make(map[UUID]session)
}

func login(platformType, platformID, IDToken string) error {
	if err := auth.Verify(platformType, IDToken); err != nil {
		return err
	}

	sess := SessionManager.SessionStart()
	sess.Set("plateformType", platformType)
	sess.Set("platformID", platformID)

	return nil
}

func checkHMAC(m, mHMAC, key []byte) (bool, error) {
	h := hmac.New(sha256.New, key)
	h.Write(m)
	return hmac.Equal(mHMAC, h.Sum(nil)), nil
}
