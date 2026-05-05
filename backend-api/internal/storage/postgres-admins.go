package storage

import "time"

type (
	Admin struct {
		Id           string
		Login        string
		PasswordHash string
		CreatedAt    time.Time
	}

	Session struct {
		AdminId   string
		Token     string
		ExpiresAt time.Time
		CreatedAt time.Time
	}
)

func (p *Postgres) CreateAdmin(login string, hash string) error {
	_, err := p.db.Exec("INSERT INTO admins(login, password_hash) VALUES ($1, $2)", login, hash)
	if err != nil {
		p.logg.Error("create admin", "type_log", "event", "error", err, "login", login)
		return err
	}

	p.logg.Info("create admin", "type_log", "event", "login", login)
	return nil
}

func (p *Postgres) DropAdmin(login string) error {
	_, err := p.db.Exec("DELETE FROM admins WHERE login = $1", login)
	if err != nil {
		p.logg.Error("drop admin", "type_log", "event", "error", err, "login", login)
		return err
	}

	p.logg.Info("drop admin", "type_log", "event", "login", login)
	return nil
}

func (p *Postgres) SearchAdmin(login string) (*Admin, error) {
	var id, passwordHash string
	var createdAt time.Time
	err := p.db.QueryRow("SELECT id, password_hash, created_at FROM admins WHERE login = $1", login).Scan(&id, &passwordHash, &createdAt)
	if err != nil {
		p.logg.Error("query admin", "type_log", "event", "error", err, "login", login)
		return nil, err
	}

	p.logg.Info("query admin", "type_log", "event", "login", login)
	return &Admin{
		Id:           id,
		Login:        login,
		PasswordHash: passwordHash,
		CreatedAt:    createdAt,
	}, nil
}

func (p *Postgres) SearchAdminById(id string) (*Admin, error) {
	var createdAt time.Time
	var login, password_hash string
	err := p.db.QueryRow("SELECT login, password_hash, created_at FROM admins WHERE id = $1", id).Scan(&login, &password_hash, &createdAt)
	if err != nil {
		p.logg.Error("query admin", "type_log", "event", "error", err, "id", id)
		return nil, err
	}

	p.logg.Info("query admin", "type_log", "event", "id", id)
	return &Admin{Id: id, Login: login, PasswordHash: password_hash, CreatedAt: createdAt}, nil
}

func (p *Postgres) CreateSession(adminId string, token string, expiresAt time.Time) (string, error) {
	var id string
	err := p.db.QueryRow("INSERT INTO admin_sessions(admin_id, token, expires_at) VALUES ($1, $2, $3) RETURNING id", adminId, token, expiresAt).Scan(&id)
	if err != nil {
		p.logg.Error("create session", "type_log", "event", "error", err, "admin", adminId)
		return "", err
	}

	p.logg.Info("create session", "type_log", "event", "admin", adminId, "session", id)
	return id, nil
}

func (p *Postgres) SearchSession(token string) (*Session, error) {
	var adminId string
	var createdAt time.Time
	var expiresAt time.Time

	err := p.db.QueryRow("SELECT admin_id, expires_at, created_at FROM admin_sessions WHERE token = $1", token).Scan(&adminId, &expiresAt, &createdAt)
	if err != nil {
		p.logg.Error("query session", "type_log", "event", "error", err)
		return nil, err
	}

	p.logg.Info("query session", "type_log", "event")
	return &Session{
		AdminId:   adminId,
		Token:     token,
		ExpiresAt: expiresAt,
		CreatedAt: createdAt,
	}, nil
}

func (p *Postgres) DropSession(token string) error {
	_, err := p.db.Exec("DELETE FROM admin_sessions WHERE token = $1", token)
	if err != nil {
		p.logg.Error("drop session", "type_log", "event", "error", err)
		return err
	}

	p.logg.Info("drop session", "type_log", "event")
	return nil
}
