package pelican

import "database/sql"

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
func (m *Module) Start() { m.Runner.Start() }
func (m *Module) Stop()  { m.Runner.Stop(); m.Generator.Close() }
