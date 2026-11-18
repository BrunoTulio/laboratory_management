package registry

import (
	"database/sql"
	"fmt"
	"strings"
)

// Item represents a simple named registry entry (teacher, class, or subject).
type Item struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// ClassWithSubjects represents an active class and its linked subjects.
type ClassWithSubjects struct {
	ID           int
	Name         string
	SubjectIDs   map[int]bool
	SubjectNames []string
}

// Repository handles persistence for school registry entities (teachers, classes, subjects).
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new registry repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// ListTeachers returns all active teachers ordered by name.
func (r *Repository) ListTeachers() ([]Item, error) {
	return r.listItems("SELECT id, name FROM teachers WHERE status = 'ACTIVE' ORDER BY name ASC")
}

// ListClasses returns all active classes ordered by name.
func (r *Repository) ListClasses() ([]Item, error) {
	return r.listItems("SELECT id, name FROM classes WHERE status = 'ACTIVE' ORDER BY name ASC")
}

// ListSubjects returns all active subjects ordered by name.
func (r *Repository) ListSubjects() ([]Item, error) {
	return r.listItems("SELECT id, name FROM subjects WHERE status = 'ACTIVE' ORDER BY name ASC")
}

// AddTeacher inserts a new teacher by name if not empty.
func (r *Repository) AddTeacher(name string) error {
	return r.insertNamed("INSERT INTO teachers (name, status) VALUES (?, 'ACTIVE') ON CONFLICT (name) DO NOTHING", name)
}

// AddClass inserts a new class by name if not empty.
func (r *Repository) AddClass(name string) error {
	return r.insertNamed("INSERT INTO classes (name, status) VALUES (?, 'ACTIVE')", name)
}

// AddSubject inserts a new subject by name if not empty.
func (r *Repository) AddSubject(name string) error {
	return r.insertNamed("INSERT INTO subjects (name, status) VALUES (?, 'ACTIVE') ON CONFLICT (name) DO NOTHING", name)
}

// DeleteTeacher removes a teacher by ID.
func (r *Repository) DeleteTeacher(id int) error {
	_, err := r.db.Exec("DELETE FROM teachers WHERE id = ?", id)
	return err
}

// DeleteClass marks a class as inactive by ID to preserve booking references.
func (r *Repository) DeleteClass(id int) error {
	_, err := r.db.Exec("UPDATE classes SET status = 'INACTIVE', updated_at = CURRENT_TIMESTAMP WHERE id = ?", id)
	return err
}

// DeleteSubject removes a subject by ID.
func (r *Repository) DeleteSubject(id int) error {
	_, err := r.db.Exec("DELETE FROM subjects WHERE id = ?", id)
	return err
}

// UpdateTeacher renames a teacher by ID.
func (r *Repository) UpdateTeacher(id int, name string) error {
	return r.updateNamed("UPDATE teachers SET name = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", id, name)
}

// UpdateClass renames a class by ID.
func (r *Repository) UpdateClass(id int, name string) error {
	return r.updateNamed("UPDATE classes SET name = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", id, name)
}

// UpdateSubject renames a subject by ID.
func (r *Repository) UpdateSubject(id int, name string) error {
	return r.updateNamed("UPDATE subjects SET name = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", id, name)
}

func (r *Repository) updateNamed(query string, id int, name string) error {
	trimmed := strings.TrimSpace(name)
	if id <= 0 || trimmed == "" {
		return fmt.Errorf("valid id and name are required")
	}
	if _, err := r.db.Exec(query, trimmed, id); err != nil {
		return fmt.Errorf("error updating registry item: %w", err)
	}
	return nil
}

// SetClassSubjects replaces the linked subjects for a given class.
func (r *Repository) SetClassSubjects(classID int, subjectIDs []int) error {
	if classID <= 0 {
		return fmt.Errorf("valid class_id is required")
	}
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM class_subjects WHERE class_id = ?", classID); err != nil {
		return err
	}
	for _, sid := range subjectIDs {
		if sid <= 0 {
			continue
		}
		if _, err := tx.Exec("INSERT INTO class_subjects (class_id, subject_id) VALUES (?, ?) ON CONFLICT DO NOTHING", classID, sid); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// ListClassesWithSubjects returns all active classes along with their linked subjects.
func (r *Repository) ListClassesWithSubjects() ([]ClassWithSubjects, error) {
	classes, err := r.ListClasses()
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(`
		SELECT cs.class_id, s.id, s.name
		FROM class_subjects cs
		INNER JOIN subjects s ON cs.subject_id = s.id
		WHERE s.status = 'ACTIVE'
		ORDER BY s.name ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("error listing class subjects: %w", err)
	}
	defer rows.Close()

	idMap := make(map[int]map[int]bool)
	nameMap := make(map[int][]string)
	for rows.Next() {
		var cid, sid int
		var sname string
		if err := rows.Scan(&cid, &sid, &sname); err != nil {
			return nil, err
		}
		if idMap[cid] == nil {
			idMap[cid] = make(map[int]bool)
		}
		idMap[cid][sid] = true
		nameMap[cid] = append(nameMap[cid], sname)
	}

	out := make([]ClassWithSubjects, 0, len(classes))
	for _, c := range classes {
		ids := idMap[c.ID]
		if ids == nil {
			ids = make(map[int]bool)
		}
		out = append(out, ClassWithSubjects{
			ID:           c.ID,
			Name:         c.Name,
			SubjectIDs:   ids,
			SubjectNames: nameMap[c.ID],
		})
	}
	return out, nil
}

// GetClassSubjectsMap returns a map of class name -> linked subject names.
func (r *Repository) GetClassSubjectsMap() (map[string][]string, error) {
	items, err := r.ListClassesWithSubjects()
	if err != nil {
		return nil, err
	}
	res := make(map[string][]string, len(items))
	for _, item := range items {
		res[item.Name] = item.SubjectNames
	}
	return res, nil
}

func (r *Repository) listItems(query string) ([]Item, error) {
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error listing registry items: %w", err)
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.ID, &item.Name); err != nil {
			return nil, fmt.Errorf("error scanning registry item: %w", err)
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *Repository) insertNamed(query, name string) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return fmt.Errorf("name is required")
	}
	if _, err := r.db.Exec(query, trimmed); err != nil {
		return fmt.Errorf("error inserting registry item: %w", err)
	}
	return nil
}
