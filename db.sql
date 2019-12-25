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
    "author"  text          not null,
    "forum"   text          not null,
    "message" text          not null,
    "votes"   int default 0 not null,
    "created" timestamptz   not null
);
create index on "thread" ("slug");
create index on "thread" ("created", "forum");
create index on "thread" ("forum", "author");

create function inc_forum_thread() returns trigger as
$$
begin
    update forum set threads = threads + 1 where slug=NEW.forum;
    return NEW;
end;
$$ language plpgsql;

create trigger thread_insert
    after insert
    on thread
    for each row
execute procedure inc_forum_thread();


create table "post"
(
    "id"       serial primary key,
    "parent"   int         not null,
    "path"     text        not null default '',
    "author"   text        not null,
    "forum"    text        not null,
    "thread"   int         not null,
    "message"  text        not null,
    "isEdited" bool        not null default false,
    "created"  timestamptz not null
);
create index on "post" ("thread");
create index on "post" ("path");
create index on "post" ("forum", "author");

create function inc_forum_post() returns trigger as
$$
begin
    update forum set posts = posts + 1 where slug=NEW.forum;
    return NEW;
end;
$$ language plpgsql;

create trigger post_insert
    after insert
    on post
    for each row
execute procedure inc_forum_post();


create table "vote"
(
    "id"       serial primary key,
    "thread"   int  not null,
    "nickname" text not null,
    "voice"    int  not null
);
create index on "vote" ("thread", "nickname");


create table "forum_user"
(
    "forum" text not null,
    "user"  text not null
);
create unique index on "forum_user" ("user", "forum");

create function add_forum_user() returns trigger as
$$
begin
    insert into forum_user (forum, "user") values (NEW.forum, NEW.author) on conflict do nothing;
    return NEW;
end;
$$ language plpgsql;

create trigger forum_user
    after insert
    on post
    for each row
execute procedure add_forum_user();
create trigger forum_user
    after insert
    on thread
    for each row
execute procedure add_forum_user();
