package config

import (
	"context"
	"log"
	"sync/atomic"
	"time"

	"secure-communication-ltd/backend/internal/services"

	"github.com/fsnotify/fsnotify"
)

var currentPolicy atomic.Value

func InitRuntimePolicy(path string) error {
	p, err := LoadPasswordPolicy(path)
	if err != nil {

		log.Printf("[policy] init: %v", err)
	}
	currentPolicy.Store(p)
	log.Printf("[policy] loaded (min=%d hist=%d upper=%v lower=%v digit=%v special=%v)",
		p.MinLength, p.History, p.RequireUpper, p.RequireLower, p.RequireDigit, p.RequireSpecial)
	return nil
}

func GetPolicy() services.PasswordPolicy {
	v := currentPolicy.Load()
	if v == nil {
		return services.DefaultPolicy()
	}
	return v.(services.PasswordPolicy)
}

func WatchPolicy(ctx context.Context, path string) error {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	if err := w.Add(path); err != nil {
		return err
	}

	go func() {
		defer w.Close()
		var timer *time.Timer

		refresh := func() {
			p, err := LoadPasswordPolicy(path)
			if err != nil {
				log.Printf("[policy] reload error: %v", err)
			}
			currentPolicy.Store(p)
			log.Printf("[policy] reloaded (min=%d hist=%d ...)", p.MinLength, p.History)
		}

		for {
			select {
			case <-ctx.Done():
				return
			case ev := <-w.Events:
				if ev.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename|fsnotify.Chmod) != 0 {
					if timer != nil {
						timer.Stop()
					}
					timer = time.AfterFunc(250*time.Millisecond, refresh) // debounce
				}
			case err := <-w.Errors:
				log.Printf("[policy] watcher error: %v", err)
			}
		}
	}()
	return nil
}
