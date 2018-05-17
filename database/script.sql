drop table if exists roles cascade;
drop table if exists accounts cascade;
drop table if exists statusbook cascade;
drop table if exists users cascade;
drop table if exists tasks cascade;
drop table if exists books cascade;


create table users(
	userid serial not null primary key,
	role  varchar(10) not null,
	login varchar(10) not null,
	password varchar(128) not null,
	firstname varchar(10) not null,
	lastname varchar(10) not null,
	position  varchar(20) not null
);

create table accounts(
	accountid serial not null primary key,
	userid int not null references users(userid),
	accountnumber varchar(255) not null,
	balance money 
);

create table tasks(
	taskid serial not null primary key,
	title varchar(20) not null,
	date date not null,
	description text not null,
	budget money not null,
	customerid int not null references users(userid)
);

create table books(
	bookid serial not null primary key,
	status varchar(10) not null,
	taskid int not null references tasks(taskid),
	customerid int not null references users(userid), 
	freelancerid int references users(userid)
);
