package pelican

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"
)

type runStore interface {
	Claim(context.Context, time.Time) (*Claim, error)
	ClaimActive(context.Context, *Claim) (bool, error)
	Finish(context.Context, *Claim, *Generation, string, string, string, time.Duration) error
	Cleanup(context.Context) error
}
type codeGenerator interface {
	Generate(context.Context, string, string, int) (*Generation, error)
}

type Runner struct {
	repo            runStore
	generator       codeGenerator
	encryptor       Encryptor
	encryptionReady bool
	startOnce       sync.Once
	stopOnce        sync.Once
	ctx             context.Context
	cancel          context.CancelFunc
	wg              sync.WaitGroup
	jobs            chan *Claim
}

func NewRunner(repo runStore, generator codeGenerator, encryptor Encryptor, encryptionReady bool) *Runner {
	ctx, cancel := context.WithCancel(context.Background())
	return &Runner{repo: repo, generator: generator, encryptor: encryptor, encryptionReady: encryptionReady, ctx: ctx, cancel: cancel, jobs: make(chan *Claim)}
}

func (r *Runner) Start() {
	r.startOnce.Do(func() {
		r.wg.Add(2)
		go func() { defer r.wg.Done(); r.scanLoop() }()
		go func() {
			defer r.wg.Done()
			for {
				select {
				case <-r.ctx.Done():
					return
				case job := <-r.jobs:
					r.execute(job)
				}
			}
		}()
	})
}
func (r *Runner) Stop() { r.stopOnce.Do(func() { r.cancel(); r.wg.Wait() }) }

func (r *Runner) scanLoop() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	var lastCleanup time.Time
	for {
		if r.ctx.Err() != nil {
			return
		}
		ctx, cancel := context.WithTimeout(r.ctx, 10*time.Second)
		now := time.Now().UTC()
		if now.Sub(lastCleanup) >= time.Hour {
			if err := r.repo.Cleanup(ctx); err == nil {
				lastCleanup = now
			}
		}
		claim, err := r.repo.Claim(ctx, now)
		cancel()
		if err != nil && r.ctx.Err() == nil {
			slog.Warn("pelican: schedule scan failed")
		}
		if claim != nil {
			// One worker and no backlog: an hour is never delayed behind old work.
			select {
			case r.jobs <- claim:
			case <-r.ctx.Done():
				r.finish(claim, nil, "", "interrupted", "shutdown", 0)
				return
			case <-time.After(time.Second):
				r.finish(claim, nil, "", "skipped", "worker_busy", 0)
			}
		}
		select {
		case <-r.ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (r *Runner) execute(claim *Claim) {
	started := time.Now()
	ctx, cancel := context.WithTimeout(r.ctx, time.Duration(claim.Config.TimeoutSeconds)*time.Second)
	defer cancel()
	if !r.encryptionReady {
		r.finish(claim, nil, "", "failed", "encryption_not_configured", 0)
		return
	}
	key, err := r.encryptor.Decrypt(claim.Config.EncryptedKey)
	if err != nil || key == "" {
		r.finish(claim, nil, "", "failed", "key_unavailable", 0)
		return
	}
	active, err := r.repo.ClaimActive(ctx, claim)
	if err != nil || !active || ctx.Err() != nil {
		r.finish(claim, nil, "", "interrupted", "claim_unavailable", 0)
		return
	}
	result, err := r.generator.Generate(ctx, key, claim.Prompt, claim.Config.MaxOutputTokens)
	status, code, preview := "succeeded", "", ""
	if err != nil {
		status, code = "failed", "request_failed"
		var ge *generationError
		if errors.As(err, &ge) {
			status, code = ge.status, ge.code
		}
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			code = "timeout"
		}
		if r.ctx.Err() != nil {
			status, code = "interrupted", "shutdown"
		}
	} else {
		preview, err = preparePreview(result.Text)
		if err != nil {
			status, code = "preview_blocked", "unsafe_or_unsupported_output"
		}
	}
	r.finish(claim, result, preview, status, code, time.Since(started))
}

func (r *Runner) finish(c *Claim, result *Generation, preview, status, code string, elapsed time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := r.repo.Finish(ctx, c, result, preview, status, code, elapsed); err != nil {
		slog.Warn("pelican: unable to save run", "run_id", c.ID)
	} else if status != "succeeded" {
		slog.Info("pelican: hourly generation skipped", "run_id", c.ID, "code", code)
	}
}
