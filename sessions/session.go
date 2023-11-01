package sessions

import (
	"encoding/base64"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type Session interface {
	Set(key string, value interface{}) error
	Get(key string) interface{}
	Delete(key string) error
	SessionID() string
}

type Provider interface {
	SessionInit(sid string) (Session, error)
	SessionRead(sid string) (Session, error)
	SessionDestory(sid string) error
	SessionGC(maxLifeTime int64)
	SessionUpdate(sid string) error
}

var provides = make(map[string]Provider)

func Register(name string, provide Provider) {
	if provide == nil {
		panic("session: Register provide is nil")
	}
	if _, ok := provides[name]; ok {
		panic("session: the provide named " + name + "have been registed")
	}
	fmt.Println("provide Register...")
	provides[name] = provide
}

type SessionManager struct {
	cookieName  string
	lock        sync.Mutex
	provider    Provider
	maxLifeTime int64
}

func NewManager(providerName, cookieName string, maxLifeTime int64) (*SessionManager, error) {
	provider, ok := provides[providerName]
	if !ok {
		return nil, fmt.Errorf("session: unknown provide %q (forgotten import?)", providerName)
	}
	fmt.Println("Session Manager...")
	return &SessionManager{cookieName: cookieName, provider: provider, maxLifeTime: maxLifeTime}, nil
}

func (sm *SessionManager) SessionID() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}

func (sm *SessionManager) SessionStart(w http.ResponseWriter, r *http.Request) (session Session) {
	sm.lock.Lock()
	defer sm.lock.Unlock()
	cookie, err := r.Cookie(sm.cookieName)
	if err != nil || cookie.Value == "" {
		// 没有生成过session
		sid := sm.SessionID()
		session, _ := sm.provider.SessionInit(sid)
		cookie := http.Cookie{Name: sm.cookieName, Value: url.QueryEscape(sid), Path: "/", HttpOnly: true}
		http.SetCookie(w, &cookie)
		// 以下方法也能设置cookie
		//w.Header().Set("Set-Cookie", cookie.String())
		//w.Header().Add("Set-Cookie", cookie.String())
		return session
	} else {
		sid, _ := url.QueryUnescape(cookie.Value)
		session, _ := sm.provider.SessionRead(sid)
		return session
	}
}

func (sm *SessionManager) SessionRead(w http.ResponseWriter, r *http.Request) (session Session, err error) {
	cookie, errs := r.Cookie(sm.cookieName)
	if errs != nil || cookie.Value == "" {
		return nil, errors.New("cookie not found")
	} else {
		sid, _ := url.QueryUnescape(cookie.Value)
		session, _ := sm.provider.SessionRead(sid)
		return session, nil
	}

}

func (sm *SessionManager) SessionDestory(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sm.cookieName)
	if cookie.Value == "" || err != nil {
		return
	} else {
		sm.lock.Lock()
		defer sm.lock.Unlock()
		sid, _ := url.QueryUnescape(cookie.Value)
		sm.provider.SessionDestory(sid)
		expiration := time.Now()
		cookie := http.Cookie{Name: sm.cookieName, Path: "/", HttpOnly: true, Expires: expiration, MaxAge: -1}
		http.SetCookie(w, &cookie)
	}
}

func (sm *SessionManager) GC() {
	sm.lock.Lock()
	defer sm.lock.Unlock()
	sm.provider.SessionGC(sm.maxLifeTime)
	time.AfterFunc(time.Duration(sm.maxLifeTime), func() { sm.GC() })
}
