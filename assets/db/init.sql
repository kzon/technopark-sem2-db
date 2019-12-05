create extension if not exists citext;

create table "user"
(
    "id"       serial primary key,
    "nickname" citext  not null unique,
    "email"    citext  not null unique,
    "fullname" varchar not null,
    "about"    varchar not null default ''
);

create table "forum"
(
    "id"      serial primary key,
    "title"   varchar       not null,
    "user"    citext        not null,
    "slug"    citext        not null unique,
    "posts"   int default 0 not null,
    "threads" int default 0 not null
);

create table "thread"
(
    "id"      serial primary key,
    "title"   varchar                   not null,
    "author"  citext                    not null,
    "forum"   citext                    not null,
    "message" text                      not null,
    "votes"   int         default 0     not null,
    "slug"    citext                    not null,
    "created" timestamptz default now() not null
);

create table "post"
(
    "id"       serial primary key,
    "parent"   int                       not null,
    "author"   citext                    not null,
    "forum"    citext                    not null,
    "thread"   int                       not null,
    "message"  text                      not null,
    "isEdited" bool        default false not null,
    "created"  timestamptz default now() not null
);
