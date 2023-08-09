// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package monitors

import (
	"time"

	"github.com/elastic/elastic-agent-libs/logp"
)

type signalWait struct {
	count   int // number of potential 'alive' signals
	signals chan struct{}
}

type signaler func()

func NewSignalWait() *signalWait {
	return &signalWait{
		signals: make(chan struct{}, 1),
	}
}

func (s *signalWait) Wait() {
	if s.count == 0 {
		return
	}

	<-s.signals
	s.count--
}

func (s *signalWait) Add(fn signaler) {
	s.count++
	go func() {
		fn()
		var v struct{}
		s.signals <- v
	}()
}

func (s *signalWait) AddChan(c <-chan struct{}) {
	s.Add(WaitChannel(c))
}

func (s *signalWait) AddTimer(t *time.Timer) {
	s.Add(WaitTimer(t))
}

func (s *signalWait) AddTimeout(d time.Duration) {
	s.Add(WaitDuration(d))
}

func (s *signalWait) Signal() {
	s.Add(func() {})
}

func WaitChannel(c <-chan struct{}) signaler {
	return func() { <-c }
}

func WaitTimer(t *time.Timer) signaler {
	return func() { <-t.C }
}

func WaitDuration(d time.Duration) signaler {
	return WaitTimer(time.NewTimer(d))
}

func WithLog(s signaler, msg string) signaler {
	return func() {
		s()
		logp.L().Infof("%v", msg)
	}
}