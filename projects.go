package main

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"errors"
	"log"
	"time"
)

// Project is a project associated with a user
type Project struct {
	// ID the project identifier
	ID int `db:"id" json:"id"`

	// UserID the ID of the project owner
	UserID int `db:"user_id" json:"userId"`

	// Name the project public name
	Name string `db:"name" json:"name"`

	// Slug what goes in the url when refering to this project. Unique user-wide but not globally.
	Slug string `db:"slug" json:"slug"`

	// DateStart the date at which the project was started
	DateStart time.Time `db:"date_start" json:"dateStart"`

	// DateEnd the date at which the project must end
	DateEnd time.Time `db:"date_end" json:"dateEnd"`

	// WordCountStart the initial word count for this writing project
	WordCountStart int `db:"word_count_start" json:"wordCountStart"`

	// WordCountGoal the goal word count for this writing project
	WordCountGoal int `db:"word_count_goal" json:"wordCountGoal"`

	// TODO: put WordCountStart & WordCountGoal in a Details field
	// TODO: add a Settings field with for instance which stats to display, privacy settings

	// Sprints the sprints on this project
	Sprints []*Sprint `json:"sprints"`
}

// FetchProjects fetches a user’s projects and returns an error or nil on success
func (u *User) FetchProjects() error {
	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return err
	}
	defer db.Close()

	rows, err := db.Queryx("select * from autochrone.projects where user_id = $1 order by name", u.ID)
	if err != nil {
		return err
	}
	u.Projects = []*Project{}
	for rows.Next() {
		p := &Project{}
		if err := rows.StructScan(p); err != nil {
			return err
		}

		u.Projects = append(u.Projects, p)
	}

	return nil
}

// GetProjectByID returns the project with the given ID and nil or nil and an error
func GetProjectByID(id int) (*Project, error) {
	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	row := db.QueryRowx("select * from autochrone.projects where id = $1", id)
	p := &Project{}
	if err := row.StructScan(p); err != nil {
		return nil, err
	}

	return p, nil
}

// GetProjectBySlug retrieves the project with the given slug belonging to the current user, and a potential error value
func (u *User) GetProjectBySlug(slug string) (*Project, error) {
	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	p := &Project{}
	if err := db.Get(p, "select * from autochrone.projects where user_id = $1 and slug = $2", u.ID, slug); err != nil {
		return nil, err
	}

	return p, nil
}

// NewProject creates a projects for the given user, inserts it in the database and returns it alongside a potential error.
func (u *User) NewProject(name, slug string, dateStart, dateEnd time.Time, wordCountStart, wordCountGoal int) (*Project, error) {
	p := &Project{
		UserID:         u.ID,
		Name:           name,
		Slug:           slug,
		DateStart:      dateStart,
		DateEnd:        dateEnd,
		WordCountStart: wordCountStart,
		WordCountGoal:  wordCountGoal,
	}

	log.Print(p)

	if p.Name == "" || p.Slug == "" || p.DateStart.Before(time.Now().Truncate(time.Hour*time.Duration(24))) || p.DateEnd.Before(p.DateStart) || p.WordCountStart < 0 || p.WordCountGoal < p.WordCountStart {
		return nil, errors.New("NewProject: invalid data")
	}

	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	row := db.QueryRowx(`
		insert into autochrone.projects(
			user_id, name, slug, date_start, date_end, word_count_start, word_count_goal
		) values ($1, $2, $3, $4, $5, $6, $7)
		returning id
	`, p.UserID, p.Name, p.Slug, p.DateStart, p.DateEnd, p.WordCountStart, p.WordCountGoal)
	if err := row.Scan(&(p.ID)); err != nil {
		return nil, err
	}

	return p, nil
}

// Update saves an existing project in the database and returns a potential error
func (p *Project) Update() error {
	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Queryx(`update autochrone.projects
		set (user_id, name, slug, date_start, date_end, word_count_start, word_count_goal)
		= ($1, $2, $3, $4, $5, $6, $7)
		where id = $8`, p.UserID, p.Name, p.Slug, p.DateStart, p.DateEnd, p.WordCountStart, p.WordCountGoal, p.ID)
	if err != nil {
		return err
	}

	return nil
}

// Delete deletes a project from the database along with all of the sprints on it
func (p *Project) Delete() error {
	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Queryx("delete from autochrone.sprints where project_id = $1", p.ID)
	if err != nil {
		return err
	}
	_, err = db.Queryx("delete from autochrone.projects where id = $1", p.ID)
	if err != nil {
		return err
	}

	return nil
}
