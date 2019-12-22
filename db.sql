create extension if not exists citext;

create table "user"
(
    "id"       serial,
    "nickname" citext not null primary key,
    "email"    citext not null unique,
    "fullname" text   not null,
    "about"    text   not null default ''
);

create table "forum"
(
    "id"      serial,
    "slug"    citext        not null primary key,
    "title"   text          not null,
    "user"    citext        not null,
    "posts"   int default 0 not null,
    "threads" int default 0 not null
);

create table "thread"
(
    "id"      serial primary key,
    "slug"    citext        not null,
    "title"   text          not null,
    "author"  citext        not null,
    "forum"   citext        not null,
    "message" text          not null,
    "votes"   int default 0 not null,
    "created" timestamptz   not null
);
create index on "thread" ("slug");
create index on "thread" ("forum", "author");


create table "post"
(
    "id"       serial primary key,
    "parent"   int         not null,
    "path"     text        not null default '',
    "author"   citext      not null,
    "forum"    citext      not null,
    "thread"   int         not null,
    "message"  text        not null,
    "isEdited" bool        not null default false,
    "created"  timestamptz not null
);
create index on "post" ("thread");
create index on "post" ("parent");
create index on "post" ("path");
create index on "post" ("created");
create index on "post" ("forum", "author");


create table "vote"
(
    "id"       serial primary key,
    "thread"   int    not null,
    "nickname" citext not null,
    "voice"    int    not null
);
create index on "vote" ("thread", "nickname");