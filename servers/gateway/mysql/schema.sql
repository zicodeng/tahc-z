-- Schema for User database.
create table if not exists user
(
    id char(64) primary key not null,
    email varchar(64) not null,
    passhash binary(64) not null,
    username  varchar(64) not null,
    firstname varchar(64) not null,
    lastname varchar(64) not null,
    photourl varchar(128) not null
)
