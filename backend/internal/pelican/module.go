package pelican

import (
	"context"
	"database/sql"
	"log/slog"
	"time"
)

type Module struct {
	Repo            *Repository
	Runner          *Runner
	Generator       *Generator
	Encryptor       Encryptor
	EncryptionReady bool
}

func NewModule(db *sql.DB, encryptor Encryptor, encryptionReady bool, endpoint string) *Module {
	repo := NewRepository(db)
	generator := NewGenerator(endpoint)
	return &Module{Repo: repo, Runner: NewRunner(repo, generator, encryptor, encryptionReady), Generator: generator, Encryptor: encryptor, EncryptionReady: encryptionReady}
}
func (m *Module) Start() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	if err := m.Repo.MigrateArtifacts(ctx); err != nil {
		slog.Error("pelican: artifact migration failed; generation stopped", "error", err)
		m.Runner.Stop()
		return
	}
	m.Runner.Start()
}
func (m *Module) Stop() { m.Runner.Stop(); m.Generator.Close() }
