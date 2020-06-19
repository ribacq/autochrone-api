create schema if not exists autochrone;
set schema 'autochrone';

-- users
create table if not exists
users (
	id serial primary key,
	username varchar(32) unique not null,
	password_hash varchar not null,
	password_salt varchar not null
);

-- projects
create table if not exists
projects (
	id serial primary key,
	user_id int not null references users(id),
	name varchar not null,
	slug varchar(32) not null,
	date_start date not null,
	date_end date not null,
	word_count_start int not null,
	word_count_goal int not null,
	unique (user_id, slug)
);

-- sprints
create table if not exists
sprints (
	id serial primary key,
	slug varchar(128) unique not null,
	project_id int not null references projects(id),
	time_start timestamp not null,
	duration int not null, --minutes
	break int not null default 0,
	word_count int not null,
	is_milestone boolean not null,
	comment varchar(1000) not null
);

-- host_sprints
create table if not exists
host_sprints (
	host_sprint_id int primary key references sprints(id),
	invite_slug varchar(128) unique not null,
	comment varchar(1000) not null
);

-- guest_sprints
create table if not exists
guest_sprints (
	guest_sprint_id int primary key references sprints(id),
	host_sprint_id int not null references host_sprints(host_sprint_id),
	check (guest_sprint_id != host_sprint_id)
);

-- sprints_with_details
create or replace view sprints_with_details as select
	sprints.*,
	coalesce(host_sprints.invite_slug, '') invite_slug,
	coalesce(host_sprints.comment, '') invite_comment,
	projects.slug project_slug,
	users.username
	from autochrone.sprints
		inner join autochrone.projects on sprints.project_id = projects.id
		inner join autochrone.users on projects.user_id = users.id
		left outer join host_sprints on sprints.id = host_sprints.host_sprint_id;
