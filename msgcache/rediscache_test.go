/*
 * Copyright 2013 Nan Deng
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package msgcache

import (
	"crypto/rand"
	"github.com/garyburd/redigo/redis"
	"github.com/uniqush/uniqush-conn/proto"
	"io"
	"testing"
	"fmt"
	"time"
)

func randomMessage() *proto.Message {
	msg := new(proto.Message)
	msg.Body = make([]byte, 10)
	io.ReadFull(rand.Reader, msg.Body)
	msg.Header = make(map[string]string, 2)
	msg.Header["aaa"] = "hello"
	msg.Header["aa"] = "hell"
	return msg
}

func multiRandomMessage(N int) []*proto.Message {
	msgs := make([]*proto.Message, N)
	for i := 0; i < N; i++ {
		msgs[i] = randomMessage()
	}
	return msgs
}

func getCache() Cache {
	db := 1
	c, _ := redis.Dial("tcp", "localhost:6379")
	c.Do("SELECT", db)
	c.Do("FLUSHDB")
	c.Close()
	return NewRedisMessageCache("", "", db)
}

func TestSetGetPoster(t *testing.T) {
	N := 10
	msgs := multiRandomMessage(N)
	cache := getCache()
	srv := "srv"
	usr := "usr"

	keys := make([]string, N)
	ids := make([]string, N)

	for i := 0; i < N; i++ {
		keys[i] = fmt.Sprintf("%v", i)
	}
	for i, msg := range msgs {
		id, err := cache.SetPoster(srv, usr, keys[i], msg, 0 * time.Second)
		if err != nil {
			t.Errorf("Set error: %v", err)
			return
		}
		ids[i] = id
	}
	for i, msg := range msgs {
		m, err := cache.GetOrDel(srv, usr, ids[i])
		if err != nil {
			t.Errorf("Get error: %v", err)
			return
		}
		if !m.Eq(msg) {
			t.Errorf("%vth message does not same", i)
		}
	}
	for i, msg := range msgs {
		m, err := cache.GetOrDel(srv, usr, ids[i])
		if err != nil {
			t.Errorf("Get error: %v", err)
			return
		}
		if !m.Eq(msg) {
			t.Errorf("%vth message does not same", i)
		}
	}
}

func TestGetSetMail(t *testing.T) {
	N := 10
	msgs := multiRandomMessage(N)
	cache := getCache()
	srv := "srv"
	usr := "usr"

	ids := make([]string, N)

	for i, msg := range msgs {
		id, err := cache.SetMail(srv, usr, msg, 0 * time.Second)
		if err != nil {
			t.Errorf("Set error: %v", err)
			return
		}
		ids[i] = id
	}
	for i, msg := range msgs {
		m, err := cache.GetOrDel(srv, usr, ids[i])
		if err != nil {
			t.Errorf("Del error: %v", err)
			return
		}
		if !m.Eq(msg) {
			t.Errorf("%vth message does not same", i)
		}
	}
	for i, id := range ids {
		m, err := cache.GetOrDel(srv, usr, id)
		if err != nil {
			t.Errorf("Get error: %v", err)
			return
		}
		if m != nil {
			t.Errorf("%vth message should be deleted", i)
		}
	}

}

func TestGetSetMailTTL(t *testing.T) {
	N := 10
	msgs := multiRandomMessage(N)
	cache := getCache()
	srv := "srv"
	usr := "usr"

	ids := make([]string, N)

	for i, msg := range msgs {
		id, err := cache.SetMail(srv, usr, msg, 1 * time.Second)
		if err != nil {
			t.Errorf("Set error: %v", err)
			return
		}
		ids[i] = id
	}
	time.Sleep(2 * time.Second)
	for i, id := range ids {
		m, err := cache.GetOrDel(srv, usr, id)
		if err != nil {
			t.Errorf("Get error: %v", err)
			return
		}
		if m != nil {
			t.Errorf("%vth message should be deleted", i)
		}
	}
}

