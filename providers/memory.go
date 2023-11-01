package providers

import (
	"container/list"
	"errors"
	"fmt"
	"github.com/wuguojun0316/GoWebPratise/sessions"
	"sync"
	"time"
)

var pder = &Provider{list: list.New()}

type Session struct {
	sid          string                 // session id
	timeAccessed time.Time              // 最后访问时间
	value        map[string]interface{} // session存储的值
}

func (session *Session) Set(key string, value interface{}) error {
	session.value[key] = value
	pder.SessionUpdate(session.sid)
	return nil
}

func (session *Session) Get(key string) interface{} {
	if _, ok := session.value[key]; !ok {
		return nil
	}
	pder.SessionUpdate(session.sid)
	return session.value[key]
}

func (session *Session) Delete(key string) error {
	delete(session.value, key)
	pder.SessionUpdate(session.sid)
	return nil
}

func (session *Session) SessionID() string {
	return session.sid
}

type Provider struct {
	lock sync.Mutex
	eles map[string]*list.Element
	list *list.List
}

func (pder *Provider) SessionInit(sid string) (sessions.Session, error) {
	pder.lock.Lock()
	defer pder.lock.Unlock()
	if v, ok := pder.eles[sid]; ok {
		return v.Value.(*Session), nil
	}
	sesValue := make(map[string]interface{}, 0)
	session := &Session{sid: sid, timeAccessed: time.Now(), value: sesValue}
	element := pder.list.PushFront(session)
	pder.eles[sid] = element
	return session, nil
}

func (pder *Provider) SessionRead(sid string) (sessions.Session, error) {
	if v, ok := pder.eles[sid]; ok {
		return v.Value.(*Session), nil
	} else {
		ses, err := pder.SessionInit(sid)
		return ses, err
	}
}

func (pder *Provider) SessionDestory(sid string) error {
	if v, ok := pder.eles[sid]; ok {
		delete(pder.eles, sid)
		pder.list.Remove(v)
		return nil
	}
	return errors.New("销毁session出错")
}

func (pder *Provider) SessionGC(maxLifeTime int64) {
	pder.lock.Lock()
	defer pder.lock.Unlock()

	for {
		var ele *list.Element = pder.list.Back()
		if ele == nil {
			break
		}
		if (ele.Value.(*Session).timeAccessed.Unix() + maxLifeTime) < time.Now().Unix() {
			pder.list.Remove(ele)
			delete(pder.eles, ele.Value.(*Session).sid)
		} else {
			break
		}
	}
}

func (pder *Provider) SessionUpdate(sid string) error {
	pder.lock.Lock()
	defer pder.lock.Unlock()

	if ele, ok := pder.eles[sid]; ok {
		ele.Value.(*Session).timeAccessed = time.Now()
		pder.list.MoveToFront(ele)
		return nil
	}
	return errors.New("更新session出错")
}

func init() {
	fmt.Println("provide init...")
	pder.eles = make(map[string]*list.Element, 0)
	sessions.Register("memory", pder)
}
