package login

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"sync"

	"github.com/satori/go.uuid"

	"bulkytree.com/sevenhearts/auth"
)

const maxLifetime int64 = 3600

type UUID uuid.UUID

func (u UUID) String() string {
	return uuid.UUID(u).String()
}

var sessions = make(map[UUID]session)
var SessionManager *sessionManager

func init() {
	SessionManager = NewSessionManager(maxLifetime)
}

type sessionManager struct {
	lock        sync.Mutex
	maxLifetime int64
}

type session interface {
	Set(key string, value interface{}) //set session value
	Get(key string) interface{}        //get session value
	Delete(key string)                 //delete session value
	SessionKey() UUID                  //back current sessionID
}

type sessionObj struct {
	session map[string]interface{}
}

func (s sessionObj) Get(key string) interface{} {
	return s.session[key]
}

func (s sessionObj) Set(key string, value interface{}) {
	s.session[key] = value
}

func (s sessionObj) Delete(key string) {
	_, ok := s.session[key]
	if ok {
		delete(s.session, key)
	}
}

func (s sessionObj) SessionKey() UUID {
	return s.session["sessionKey"].(UUID)
}

func (manager *sessionManager) NewSessionKey() UUID {
	return UUID(uuid.NewV4())
}

func (manager *sessionManager) sessionInit(sKey UUID) session {
	s := sessionObj{}
	s.session = make(map[string]interface{})
	s.Set("sessionKey", sKey)
	sessions[sKey] = s
	return s
}

func (manager *sessionManager) SessionRead(sKey UUID) session {
	s, ok := sessions[sKey]
	if ok {
		return manager.sessionInit(sKey)
	}
	return s
}

func (manager *sessionManager) SessionDestroy(sKey UUID) {
	_, ok := sessions[sKey]
	if ok {
		delete(sessions, sKey)
	}
}

func (manager *sessionManager) SessionGC(maxLifetime int64) {

}

func (manager *sessionManager) SessionStart() (session session) {
	manager.lock.Lock()
	defer manager.lock.Unlock()
	sKey := manager.NewSessionKey()
	session = manager.sessionInit(sKey)
	fmt.Printf("session: %v, sessionKey: %v, sKey: %v", session, session.SessionKey(), sKey)
	return session
}

func NewSessionManager(maxLifetime int64) *sessionManager {
	return &sessionManager{maxLifetime: maxLifetime}
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
