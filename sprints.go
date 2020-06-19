package main

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"errors"
	"fmt"
	"time"
)

// Sprint is a sprint on a project
type Sprint struct {
	// ID the sprint ID
	ID int `db:"id" json:"id"`

	// Slug what goes in the url when refering to this project. Unique globally.
	Slug string `db:"slug" json:"slug"`

	// Username the username of the sprint’s project’s user
	Username string `db:"username" json:"username"`

	// ProjectSlug the slug of the project the sprint is for
	ProjectSlug string `db:"project_slug" json:"pslug"`

	// ProjectID the ID of the project the sprint is for
	ProjectID int `db:"project_id" json:"projectId"`

	// TimeStart the moment at which the sprint starts
	TimeStart time.Time `db:"time_start" json:"timeStart"`

	// Duration duration of the sprint in minutes
	Duration int `db:"duration" json:"duration"`

	// WordCount the word count of the writing project
	WordCount int `db:"word_count" json:"wordCount"`

	// Break the break that must follow the sprint in minutes
	Break int `db:"break" json:"break"`

	// IsMilestone whether this sprints is a milestone for the project
	IsMilestone bool `db:"is_milestone" json:"isMilestone"`

	// Comment a comment on the sprint
	Comment string `db:"comment" json:"comment"`

	// InviteSlug the invite slug, empty if the sprint is not open to guests
	InviteSlug string `db:"invite_slug" json:"inviteSlug"`

	// InviteComment the invite comment, empty if the sprint is not open to guests
	InviteComment string `db:"invite_comment" json:"inviteComment"`
}

// FetchSprints fetches the sprints on a given project, returning a potential error
func (p *Project) FetchSprints() error {
	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return err
	}
	defer db.Close()

	rows, err := db.Queryx("select * from sprints_with_details where project_id = $1 order by time_start desc", p.ID)
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
	Date time.Time `json:"date"`

	// Sprints the sprints that have this date in common
	Sprints []*Sprint `json:"sprints"`
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
	if err := db.Get(s, "select * from sprints_with_details where id = $1", id); err != nil {
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
	if err := db.Get(s, "select * from sprints_with_details where slug = $1", slug); err != nil {
		return nil, err
	}
	return s, nil
}

// GetSprintByInviteSlug returns the sprint with the given invite slug in autochrone.host_sprints sql table
func GetSprintByInviteSlug(inviteSlug string) (*Sprint, error) {
	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	s := &Sprint{}
	if err := db.Get(s, "select * from sprints_with_details where invite_slug = $1", inviteSlug); err != nil {
		return nil, err
	}
	return s, nil
}

// NewSprint adds a sprint to a project and inserts it in the database
func (p *Project) NewSprint(timeStart time.Time, duration, pomodoroBreak int) (*Sprint, error) {
	if duration < 1 || pomodoroBreak < 0 {
		return nil, errors.New("NewSprint: invalid duration or pomodoroBreak values")
	}

	s := &Sprint{
		Slug:      fmt.Sprintf("%x.%x", p.ID, timeStart.UTC().UnixNano()),
		ProjectID: p.ID,
		TimeStart: timeStart.UTC(),
		Duration:  duration,
		Break:     pomodoroBreak,
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
	`, s.Slug, s.ProjectID, s.TimeStart.UTC().Format("2006-01-02 15:04:05"), s.Duration, s.Break, 0, false, "")
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

// GetNextSprintIfExists returns a the sprint on the same project that starts at sprint.TimeEnd + sprint.Break.
// returns a pointer to sprint and a boolean set to true if it was found.
func (s *Sprint) GetNextSprintIfExists() (nextSprint *Sprint, ok bool) {
	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return nil, false
	}
	defer db.Close()

	nextSprint = &Sprint{}
	if err := db.Get(nextSprint, "select * from autochrone.sprints where project_id = $1 and time_start > $2 order by time_start asc limit 1", s.ProjectID, s.TimeEnd().UTC().Format("2006-01-02 15:04:05")); err != nil {
		return nil, false
	}

	return nextSprint, true
}

// IsSingleSprint returns true if the sprint is not in a pomodoro streak
func (s *Sprint) IsSingleSprint() bool {
	return s.Break == 0
}

// IsOpenToGuests is true if a sprint has an invite slug
func (s *Sprint) IsOpenToGuests() bool {
	return s.InviteSlug != ""
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

// OpenToGuests opens the sprint to guests, sets the public comment and returns the invite slug (and error).
func (s *Sprint) OpenToGuests(comment string) (string, error) {
	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return "", err
	}
	defer db.Close()

	inviteSlug := fmt.Sprintf("%s.%x", s.Slug, time.Now().UnixNano())
	_, err = db.Queryx("insert into autochrone.host_sprints (host_sprint_id, invite_slug, comment) values ($1, $2, $3)", s.ID, inviteSlug, comment)
	if err != nil {
		return "", err
	}

	return inviteSlug, nil
}

// GetGuestSprints get the guest sprints from the database (and error).
func (s *Sprint) GetGuestSprints() ([]*Sprint, error) {
	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Queryx(`select *
		from sprints_with_details
		inner join guest_sprints on sprints_with_details.id = guest_sprints.host_sprint_id
		where guest_sprints.host_sprint_id = $1`, s.ID)
	if err != nil {
		return nil, err
	}
	guestSprints := []*Sprint{}
	for rows.Next() {
		s := &Sprint{}
		if err := rows.StructScan(s); err != nil {
			return nil, err
		}
		guestSprints = append(guestSprints, s)
	}

	return guestSprints, nil
}

// NewGuestSprint creates a guest sprint on the given project with the model host sprint
func (p *Project) NewGuestSprint(hostSprint *Sprint) (*Sprint, error) {
	if hostSprint.Over() {
		return nil, errors.New("NewGuestSprint: host sprint is over.")
	}

	return p.NewSprint(hostSprint.TimeStart, hostSprint.Duration, hostSprint.Break)
}
