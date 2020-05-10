package main

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"fmt"
	"time"
)

// Sprint is a sprint on a project
type Sprint struct {
	// ID the sprint ID
	ID int `db:"id"`

	// Slug what goes in the url when refering to this project. Unique globally.
	Slug string `db:"slug"`

	// ProjectID the ID of the project the sprint is for
	ProjectID int `db:"project_id"`

	// TimeStart the moment at which the sprint starts
	TimeStart time.Time `db:"time_start"`

	// Duration duration of the sprint in minutes
	Duration int `db:"duration"`

	// WordCount the word count of the writing project
	WordCount int `db:"word_count"`

	// Break the break that must follow the sprint in minutes
	Break int `db:"break"`

	// IsMilestone whether this sprints is a milestone for the project
	IsMilestone bool `db:"is_milestone"`

	// Comment a comment on the sprint
	Comment string `db:"comment"`
}

// FetchSprints fetches the sprints on a given project, returning a potential error
func (p *Project) FetchSprints() error {
	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return err
	}
	defer db.Close()

	rows, err := db.Queryx("select * from autochrone.sprints where project_id = $1 order by time_start desc", p.ID)
	if err != nil {
		return err
	}
	p.Sprints = []*Sprint{}
	for rows.Next() {
		s := &Sprint{}
		if err := rows.StructScan(s); err != nil {
			return err
		}
		p.Sprints = append(p.Sprints, s)
	}

	return nil
}

// DateSprints groups sprints happening on a common date
type DateSprints struct {
	// Date the common date of the sprints
	Date time.Time

	// Sprints the sprints that have this date in common
	Sprints []*Sprint
}

// GetSprintsByDate returns the sprints for the project grouped by day
func (p *Project) GetSprintsByDate() ([]DateSprints, error) {
	var ret []DateSprints

	for _, s := range p.Sprints {
		index := -1
		for i := range ret {
			if ret[i].Date.UTC() == s.TimeStart.Truncate(24*time.Hour).UTC() {
				index = i
				break
			}
		}
		if index == -1 {
			ret = append(ret, DateSprints{
				Date:    s.TimeStart.Truncate(24 * time.Hour),
				Sprints: []*Sprint{s},
			})
		} else {
			ret[index].Sprints = append(ret[index].Sprints, s)
		}
	}

	return ret, nil
}

// GetSprintByID returns the sprint with the given ID and a potential error
func GetSprintByID(id int) (*Sprint, error) {
	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	s := &Sprint{}
	if err := db.Get(s, "select * from autochrone.sprints where id = $1", id); err != nil {
		return nil, err
	}
	return s, nil
}

// GetSprintBySlug returns the sprint with the given slug and a potential error
func GetSprintBySlug(slug string) (*Sprint, error) {
	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	s := &Sprint{}
	if err := db.Get(s, "select * from autochrone.sprints where slug = $1", slug); err != nil {
		return nil, err
	}
	return s, nil
}

// NewSprint adds a sprint to a project and inserts it in the database
func (p *Project) NewSprint(timeStart time.Time, duration int) (*Sprint, error) {
	s := &Sprint{
		Slug:      fmt.Sprintf("%xo%x", p.ID, timeStart.UTC().UnixNano()),
		ProjectID: p.ID,
		TimeStart: timeStart.UTC(),
		Duration:  duration,
	}

	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	row := db.QueryRowx(`
		insert into autochrone.sprints(
			slug, project_id, time_start, duration, break, word_count, is_milestone, comment
		) values ($1, $2, $3, $4, $5, $6, $7, $8)
		returning id
	`, s.Slug, s.ProjectID, s.TimeStart.UTC().Format("2006-01-02 15:04:05"), s.Duration, 0, 0, false, "")
	if err := row.Scan(&(s.ID)); err != nil {
		return nil, err
	}

	p.Sprints = append(p.Sprints, s)
	return s, nil
}

// Update saves an existing sprint in the database
func (s *Sprint) Update() error {
	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Queryx(`update autochrone.sprints
		set (time_start, duration, break, word_count, is_milestone, comment)
		= ($1, $2, $3, $4, $5, $6)
		where id = $7`, s.TimeStart.UTC().Format("2006-01-02 15:04:05"), s.Duration, s.Break, s.WordCount, s.IsMilestone, s.Comment, s.ID)
	if err != nil {
		return err
	}

	return nil
}

// Delete removes a sprint from the database
func (s *Sprint) Delete() error {
	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Queryx("delete from autochrone.sprints where id = $1", s.ID)
	if err != nil {
		return err
	}

	return nil
}

// TimeEnd returns the time at which the sprint ends
func (s *Sprint) TimeEnd() time.Time {
	return s.TimeStart.Add(time.Duration(s.Duration) * time.Minute)
}

// Upcoming returns true if the sprint has not yet started
func (s *Sprint) Upcoming() bool {
	return s.TimeStart.After(time.Now())
}

// Running returns true if the sprint has started and is not over
func (s *Sprint) Running() bool {
	return s.TimeEnd().After(time.Now()) && s.TimeStart.Before(time.Now())
}

// Over returns true if the sprint is over
func (s *Sprint) Over() bool {
	return s.TimeEnd().Before(time.Now())
}

// MilestoneIndex returns the number of milestones prior to this sprint plus 1.
// The sprint needs not be a milestone itself.
func (s *Sprint) MilestoneIndex() (int, error) {
	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return 0, err
	}
	defer db.Close()

	row := db.QueryRowx("select count(id) from autochrone.sprints where project_id = $1 and time_start <= $2 and is_milestone = true", s.ProjectID, s.TimeStart.UTC().Format("2006-01-02 15:04:05"))
	var i int
	if err := row.Err(); err != nil {
		return 0, err
	}
	if err := row.Scan(&i); err != nil {
		return 0, err
	}
	return i, nil
}

// PreviousMilestone returns the last milestone before this sprint or nil and an error
func (s *Sprint) PreviousMilestone() (*Sprint, error) {
	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	row := db.QueryRowx(`select id, time_start from autochrone.sprints where project_id = $1 and time_start < $2 and is_milestone = true
		union all (select -1, time_start from autochrone.sprints where project_id = $1 order by time_start limit 1)
		order by time_start desc limit 1`, s.ProjectID, s.TimeStart.UTC().Format("2006-01-02 15:04:05"))
	var id int
	if err := row.Err(); err != nil {
		return nil, err
	}

	var ts string
	if err := row.Scan(&id, &ts); err != nil {
		return nil, err
	}

	if id == -1 {
		return nil, nil
	}

	return GetSprintByID(id)
}

// MilestoneWordCount returns the number of words written since the last milestone was set
// excluding the sprint on which the last milestone was set and including the current sprint.
// The current sprint needs not be a milestone itself.
func (s *Sprint) MilestoneWordCount() (int, error) {
	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return 0, err
	}
	defer db.Close()

	previousMilestone, err := s.PreviousMilestone()
	if err != nil {
		return 0, err
	}

	var row *sqlx.Row
	if previousMilestone != nil {
		row = db.QueryRowx("select sum(word_count) from autochrone.sprints where project_id = $1 and time_start <= $2 and time_start > $3", s.ProjectID, s.TimeStart.UTC().Format("2006-01-02 15:04:05"), previousMilestone.TimeStart.UTC().Format("2006-01-02 15:04:05"))
	} else {
		row = db.QueryRowx("select sum(word_count) from autochrone.sprints where project_id = $1 and time_start <= $2", s.ProjectID, s.TimeStart.UTC().Format("2006-01-02 15:04:05"))
	}
	var wc int
	if err := row.Err(); err != nil {
		return 0, err
	}
	if err := row.Scan(&wc); err != nil {
		return 0, err
	}

	if wc == -1 {
		return s.WordCount, nil
	}

	return wc, nil
}

// MilestoneTimeSpent returns the duration spent since the last milestone was set
// excluding the sprint on which the last milestone was set and including the current sprint.
// The current sprint needs not be a milestone itself.
func (s *Sprint) MilestoneTimeSpent() (time.Duration, error) {
	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return 0, err
	}
	defer db.Close()

	previousMilestone, err := s.PreviousMilestone()
	if err != nil {
		return 0, err
	}

	var row *sqlx.Row
	if previousMilestone != nil {
		row = db.QueryRowx("select sum(duration) from autochrone.sprints where project_id = $1 and time_start <= $2 and time_start > $3", s.ProjectID, s.TimeStart.UTC().Format("2006-01-02 15:04:05"), previousMilestone.TimeStart.UTC().Format("2006-01-02 15:04:05"))
	} else {
		row = db.QueryRowx("select sum(duration) from autochrone.sprints where project_id = $1 and time_start <= $2", s.ProjectID, s.TimeStart.UTC().Format("2006-01-02 15:04:05"))
	}
	var d int
	if err := row.Err(); err != nil {
		return 0, err
	}
	if err := row.Scan(&d); err != nil {
		return 0, err
	}

	if d == -1 {
		return time.Duration(s.Duration) * time.Minute, nil
	}

	return time.Duration(d) * time.Minute, nil
}