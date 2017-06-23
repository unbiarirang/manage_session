package login

import (
	"testing"
	"time"
)

// TODO:
// 1. GC 만드세요. maxLife/defaultLife인지 분명히 하세요 @
// 2. NewSessionManager 가 있으면 대응되는 sessions (session store) 생성자도 있어야 합니다. 아니면 아예 둘다 지원하지 않는것도 방법 @
// 3. lock 처리가 빠져 있습니다. 보강 하세요. @
// 4. map get set ok처리 (nil 처리) @

// Q
// 1. sessionKey랑 expiry 맵에서 겟할 때와 타입 캐스팅할 때 에러처리 필요한가? @ 맵에서 빼고 세션 오브젝트의 필드로 바꿈
// 2. 하나의 session 안에서도 lock이 필요할까? @ (A:No)

func TestSession(t *testing.T) {
	s := SessionManager.SessionStart()
	if len(s.SessionKey()) != 16 {
		t.Fatal("session start 시 session키가 설정되어야 한다.")
	}

	intValue := 1
	s.Set("a", intValue)
	intResult, ok := s.Get("a")
	intResult = intResult.(int)
	if !ok || intResult != intValue {
		t.Error("get set int fail")
	}

	s1 := SessionManager.SessionStart()
	s1.Set("b", 2)
	s1.Delete("b")
	_, ok = s1.Get("b")
	if ok == true {
		t.Error("b 삭제는 성공해야 한다")
	}

	if len(sessions) != 2 {
		t.Errorf("sessions에는 두 개의 session이 존재햐야 한다. len: %v != 2", len(sessions))
	}

	time.Sleep(time.Second * 20)
}
