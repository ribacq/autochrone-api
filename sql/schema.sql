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

-- access_tokens
create table if not exists
access_tokens (
	id serial primary key,
	user_id int not null references users(id),
	token_type varchar not null,
	expires_on timestamp not null,
	scope varchar not null
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
-- TODO: for multi user support, divide sprints table into sprints and sprints_participants
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
	comment varchar not null
);
