-- Initial SQLite schema for Laboratory Management System

CREATE TABLE IF NOT EXISTS roles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role_id INTEGER REFERENCES roles(id),
    registration_number TEXT DEFAULT '',
    phone TEXT DEFAULT '',
    status TEXT DEFAULT 'ACTIVE',
    classes_csv TEXT DEFAULT '',
    subjects_csv TEXT DEFAULT '',
    activation_token TEXT DEFAULT '',
    activation_expires_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS laboratories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT DEFAULT '',
    location TEXT DEFAULT '',
    capacity INTEGER DEFAULT 0,
    status TEXT DEFAULT 'ACTIVE',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS classes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    course TEXT DEFAULT '',
    grade_year TEXT DEFAULT '',
    shift TEXT DEFAULT '',
    status TEXT DEFAULT 'ACTIVE',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS bookings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id),
    laboratory_id INTEGER NOT NULL REFERENCES laboratories(id),
    class_id INTEGER REFERENCES classes(id),
    booking_date DATE NOT NULL,
    start_time DATETIME NOT NULL,
    end_time DATETIME NOT NULL,
    shift TEXT DEFAULT '',
    teacher_name TEXT DEFAULT '',
    class_name TEXT DEFAULT '',
    subject_name TEXT DEFAULT '',
    reason TEXT NOT NULL,
    observation TEXT DEFAULT '',
    status TEXT DEFAULT 'PENDING',
    approved_by INTEGER REFERENCES users(id),
    approved_at DATETIME,
    rejected_by INTEGER REFERENCES users(id),
    rejected_at DATETIME,
    rejection_reason TEXT DEFAULT '',
    cancelled_by INTEGER REFERENCES users(id),
    cancelled_at DATETIME,
    cancellation_reason TEXT DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS notifications (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id),
    booking_id INTEGER REFERENCES bookings(id),
    type TEXT NOT NULL,
    title TEXT NOT NULL,
    message TEXT NOT NULL,
    is_read BOOLEAN DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS audit_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER REFERENCES users(id),
    action TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id INTEGER,
    details TEXT DEFAULT '',
    ip_address TEXT DEFAULT '',
    user_agent TEXT DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role_id);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);
CREATE INDEX IF NOT EXISTS idx_users_activation_token ON users(activation_token);

CREATE INDEX IF NOT EXISTS idx_bookings_user ON bookings(user_id);
CREATE INDEX IF NOT EXISTS idx_bookings_laboratory ON bookings(laboratory_id);
CREATE INDEX IF NOT EXISTS idx_bookings_class ON bookings(class_id);
CREATE INDEX IF NOT EXISTS idx_bookings_date ON bookings(booking_date);
CREATE INDEX IF NOT EXISTS idx_bookings_status ON bookings(status);
CREATE INDEX IF NOT EXISTS idx_bookings_datetime ON bookings(booking_date, start_time, end_time);

CREATE INDEX IF NOT EXISTS idx_notifications_user ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_read ON notifications(is_read);

CREATE INDEX IF NOT EXISTS idx_audit_user ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_entity ON audit_logs(entity_type, entity_id);

INSERT INTO roles (name, description) VALUES
('PROFESSOR', 'Professor role with booking permissions'),
('ALUNO', 'Student role with limited booking permissions'),
('DIRECAO', 'Administration role with full permissions')
ON CONFLICT (name) DO NOTHING;

INSERT INTO laboratories (id, name, description, location, capacity) VALUES
(1, 'Laboratório Principal', 'Laboratório de informática principal', 'Bloco A - Sala 101', 30)
ON CONFLICT (id) DO NOTHING;
