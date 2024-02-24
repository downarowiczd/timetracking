package domain

import (
	"database/sql"
	"errors"
	"time"
)

var (
	ErrDuplicate    = errors.New("record already exists")
	ErrNotExists    = errors.New("row not exists")
	ErrUpdateFailed = errors.New("update failed")
	ErrDeleteFailed = errors.New("delete failed")
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{
		db: db,
	}
}

func (r *SQLiteRepository) Migrate() error {
	query := `
	CREATE TABLE IF NOT EXISTS project(
		tag  VARCHAR(20) PRIMARY KEY UNIQUE,
		name VARCHAR(50) NOT NULL,
		type VARCHAR(20) NOT NULL,
		status INTEGER NOT NULL
	);

	CREATE TABLE IF NOT EXISTS record(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		projTag VARCHAR(20) NOT NULL,
		startTime DATETIME NOT NULL,
		endTime DATETIME,
		name VARCHAR(70) NOT NULL,
		billable BOOLEAN,
		note TEXT,
		status INTEGER NOT NULL
	);
	`
	_, err := r.db.Exec(query)
	return err
}

func (r *SQLiteRepository) CreateProject(project Project) (*Project, error) {
	_, err := r.db.Exec("INSERT INTO project(tag, name, type, status) values(?,?,?,?)", project.Tag, project.Name, project.Type, 0)
	if err != nil {
		return nil, err
	}

	return &project, nil
}

func (r *SQLiteRepository) CreateRecording(recording Recording) (*Recording, error) {
	if recording.StartTime.IsZero() {
		recording.StartTime = time.Now()
	}
	res, err := r.db.Exec("INSERT INTO record(projTag, startTime, name, billable, note, status) values(?,?,?,?,?,?)", recording.ProjectTag, recording.StartTime, recording.Name, recording.Billable, recording.Note, recording.Status)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	recording.ID = id

	return &recording, nil
}

func (r *SQLiteRepository) AllProjects() ([]Project, error) {
	rows, err := r.db.Query("SELECT * FROM project")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var all []Project
	for rows.Next() {
		var project Project
		if err := rows.Scan(&project.Tag, &project.Name, &project.Type, &project.Status); err != nil {
			return nil, err
		}
		all = append(all, project)
	}
	return all, nil
}

func (r *SQLiteRepository) AllActiveProjects() ([]Project, error) {
	rows, err := r.db.Query("SELECT * FROM project WHERE status = 0")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var all []Project
	for rows.Next() {
		var project Project
		if err := rows.Scan(&project.Tag, &project.Name, &project.Type, &project.Status); err != nil {
			return nil, err
		}
		all = append(all, project)
	}
	return all, nil
}

func (r *SQLiteRepository) AllRecordings() ([]Recording, error) {
	rows, err := r.db.Query("SELECT * FROM record")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var all []Recording
	for rows.Next() {
		var recording Recording
		if err := rows.Scan(&recording.ID, &recording.ProjectTag, &recording.StartTime, &recording.EndTime, &recording.Name, &recording.Billable, &recording.Note, &recording.Status); err != nil {
			return nil, err
		}
		all = append(all, recording)
	}
	return all, nil
}

func (r *SQLiteRepository) GetProjectByTag(tag string) (*Project, error) {
	row := r.db.QueryRow("SELECT * FROM project WHERE tag = ?", tag)
	
	var project Project
	if err := row.Scan(&project.Tag, &project.Name, &project.Type, &project.Status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotExists
		}
		return nil, err
	}
	return &project, nil
}

func (r *SQLiteRepository) GetRecordingsByProjectTag(tag string) ([]Recording, error) {
	rows, err := r.db.Query("SELECT * FROM record projTag = ?", tag)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var all []Recording
	for rows.Next() {
		var recording Recording
		if err := rows.Scan(&recording.ID, &recording.ProjectTag, &recording.StartTime, &recording.EndTime, &recording.Name, &recording.Billable, &recording.Note, &recording.Status); err != nil {
			return nil, err
		}
		all = append(all, recording)
	}
	return all, nil
}

func (r *SQLiteRepository) GetRecordingsByDateRange(start, end time.Time) ([]Recording, error) {
	rows, err := r.db.Query("SELECT * FROM record startTime >= ? AND endTime <= ?", start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var all []Recording
	for rows.Next() {
		var recording Recording
		if err := rows.Scan(&recording.ID, &recording.ProjectTag, &recording.StartTime, &recording.EndTime, &recording.Name, &recording.Billable, &recording.Note, &recording.Status); err != nil {
			return nil, err
		}
		all = append(all, recording)
	}
	return all, nil
}

func (r *SQLiteRepository) UpdateProject(tag string, updated Project) (*Project, error) {
	if tag == "" {
		return nil, errors.New("invalid project tag")
	}
	res, err := r.db.Exec("UPDATE project SET name = ?, type = ?, status = ? WHERE tag = ?", updated.Name, updated.Type, updated.Status, tag)
	if err != nil {
		return nil, err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}

	if rowsAffected == 0 {
		return nil, ErrUpdateFailed
	}

	return &updated, nil
}

func (r *SQLiteRepository) DeleteProject(tag string) error {
	res, err := r.db.Exec("DELETE FROM project WHERE tag = ?", tag)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrDeleteFailed
	}

	return err
}

func (r *SQLiteRepository) UpdateRecording(id int64, updated Recording) (*Recording, error) {
	if id == 0 {
		return nil, errors.New("invalid recording id")
	}
	res, err := r.db.Exec("UPDATE record SET projTag = ?, name = ?, startTime = ?, endTime = ?, note = ?, billable = ?, status = ? WHERE id = ?", updated.ProjectTag, updated.Name, updated.StartTime, updated.EndTime, updated.Note, updated.Billable, updated.Status, id)
	if err != nil {
		return nil, err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}

	if rowsAffected == 0 {
		return nil, ErrUpdateFailed
	}

	return &updated, nil
}

func (r *SQLiteRepository) DeleteRecording(id int64) error {
	res, err := r.db.Exec("DELETE FROM record WHERE id = ?", id)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrDeleteFailed
	}

	return err
}
