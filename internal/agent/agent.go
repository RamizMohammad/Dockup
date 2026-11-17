// Package agent is the remote worker: it keeps one outbound SSE connection to
// the hub (no inbound ports needed), runs jobs locally and streams logs back.
package agent

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/RamizMohammad/dockup/internal/runner"
)

type Agent struct {
	Hub, Token, DataDir, Version string

	mu     sync.Mutex
	cancel map[string]context.CancelFunc
	run    *runner.Runner
	post   *http.Client
}

func (a *Agent) Run(ctx context.Context) error {
	a.cancel = map[string]context.CancelFunc{}
	a.run = &runner.Runner{WorkDir: a.DataDir}
	a.post = &http.Client{Timeout: 20 * time.Second}
	a.Hub = strings.TrimRight(a.Hub, "/")
	backoff := time.Second
	for ctx.Err() == nil {
		start := time.Now()
		err := a.connect(ctx)
		if ctx.Err() != nil {
			break
		}
		if time.Since(start) > 30*time.Second {
			backoff = time.Second
		}
		log.Printf("agent: disconnected (%v), retrying in %s", err, backoff)
		select {
		case <-ctx.Done():
		case <-time.After(backoff):
		}
		if backoff < 30*time.Second {
			backoff *= 2
		}
	}
	a.mu.Lock()
	for _, c := range a.cancel {
		c()
	}
	a.mu.Unlock()
	return nil
}

func (a *Agent) ping() {
	h, o, d := runner.Info()
	a.mu.Lock()
	running := []string{}
	for id := range a.cancel {
		running = append(running, id)
	}
	a.mu.Unlock()
	b, _ := json.Marshal(map[string]any{"hostname": h, "os": o, "docker": d, "version": a.Version, "running": running})
	a.do("POST", "/api/agent/ping", bytes.NewReader(b), "application/json")
}

func (a *Agent) do(method, path string, body io.Reader, ct string) error {
	req, err := http.NewRequest(method, a.Hub+path, body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+a.Token)
	if ct != "" {
		req.Header.Set("Content-Type", ct)
	}
	resp, err := a.post.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	if resp.StatusCode >= 300 {
		return fmt.Errorf("hub answered %d", resp.StatusCode)
	}
	return nil
}

func (a *Agent) connect(ctx context.Context) error {
	req, _ := http.NewRequestWithContext(ctx, "GET", a.Hub+"/api/agent/stream", nil)
	req.Header.Set("Authorization", "Bearer "+a.Token)
	req.Header.Set("Accept", "text/event-stream")
	cl := &http.Client{} // no timeout: long-lived stream
	resp, err := cl.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("hub answered %d (check token)", resp.StatusCode)
	}
	log.Printf("agent: connected to %s", a.Hub)
	go a.ping()

	// heartbeat info refresh while connected
	hbCtx, stop := context.WithCancel(ctx)
	defer stop()
	go func() {
		t := time.NewTicker(30 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-hbCtx.Done():
				return
			case <-t.C:
				a.ping()
			}
		}
	}()

	sc := bufio.NewScanner(resp.Body)
	sc.Buffer(make([]byte, 64*1024), 8*1024*1024)
	var event, data string
	for sc.Scan() {
		line := sc.Text()
		switch {
		case strings.HasPrefix(line, "event: "):
			event = strings.TrimPrefix(line, "event: ")
		case strings.HasPrefix(line, "data: "):
			data = strings.TrimPrefix(line, "data: ")
		case line == "":
			if event != "" {
				a.handle(ctx, event, data)
			}
			event, data = "", ""
		}
	}
	if err := sc.Err(); err != nil {
		return err
	}
	return io.EOF
}

func (a *Agent) handle(ctx context.Context, event, data string) {
	switch event {
	case "job":
		var j runner.Job
		if json.Unmarshal([]byte(data), &j) != nil {
			return
		}
		go a.execute(ctx, j)
	case "cancel":
		var id string
		json.Unmarshal([]byte(data), &id)
		a.mu.Lock()
		if c := a.cancel[id]; c != nil {
			c()
		}
		a.mu.Unlock()
	}
}

func (a *Agent) execute(parent context.Context, j runner.Job) {
	ctx, cancel := context.WithCancel(parent)
	a.mu.Lock()
	if a.cancel[j.BuildID] != nil { // duplicate delivery
		a.mu.Unlock()
		cancel()
		return
	}
	a.cancel[j.BuildID] = cancel
	a.mu.Unlock()
	defer func() {
		cancel()
		a.mu.Lock()
		delete(a.cancel, j.BuildID)
		a.mu.Unlock()
	}()

	ls := &logSender{a: a, id: j.BuildID, flushed: make(chan struct{})}
	go ls.loop()
	res := a.run.Run(ctx, j, ls.add)
	ls.stop()

	body, _ := json.Marshal(map[string]any{"buildId": j.BuildID, "commit": res.Commit, "err": res.Err})
	for i := 0; i < 6; i++ {
		if err := a.do("POST", "/api/agent/done", bytes.NewReader(body), "application/json"); err == nil {
			return
		}
		time.Sleep(time.Duration(i+1) * 2 * time.Second)
	}
	log.Printf("agent: could not report result for build %s", j.BuildID)
}

// logSender batches console lines (200ms) so UI streaming feels live without
// one HTTP request per line.
type logSender struct {
	a       *Agent
	id      string
	mu      sync.Mutex
	buf     []string
	done    bool
	flushed chan struct{}
}

func (l *logSender) add(s string) {
	l.mu.Lock()
	l.buf = append(l.buf, s)
	l.mu.Unlock()
}

func (l *logSender) take() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	b := l.buf
	l.buf = nil
	return b
}

func (l *logSender) send(lines []string) {
	if len(lines) == 0 {
		return
	}
	body := []byte(strings.Join(lines, "\n") + "\n")
	for i := 0; i < 4; i++ {
		if l.a.do("POST", "/api/agent/log?build="+l.id, bytes.NewReader(body), "text/plain") == nil {
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func (l *logSender) loop() {
	t := time.NewTicker(200 * time.Millisecond)
	defer t.Stop()
	for range t.C {
		l.send(l.take())
		l.mu.Lock()
		d := l.done
		l.mu.Unlock()
		if d {
			l.send(l.take())
			close(l.flushed)
			return
		}
	}
}

func (l *logSender) stop() {
	l.mu.Lock()
	l.done = true
	l.mu.Unlock()
	<-l.flushed
}
