create table users (
  -- format as uuid-v7
  id text,
  name text not null,
  age int not null,
  -- weight todo
  _version int not null,
  check (length(name) <= 30)
  check (age >= 0 and age <= 255)
  primary key (id)
) strict;

create table _users_history (
  _rowid int,
  -- all nullable to save space
  user text,
  name text,
  age int,
  -- optimistic version control
  _version int not null,
  -- mask (https://simonwillison.net/2023/Apr/15/sqlite-history/)
  _mask int not null,
  check (_version >= 1)
  -- do i need to include the constraints again?
  foreign key (user) references users (id) -- cascade on delete?
) strict;

create index id_users_history_rowid on _users_history (_rowid);

create trigger users_insert_history 
after insert on users
begin
  insert into _users_history (_rowid, user, name, age, _version, _mask)
  values (
    new.rowid,
    new.id,
    new.name,
    new.age,
    1,
    15
  );
end;

create trigger users_update_history
after update on users for each row
begin
  insert into _users_history (_rowid, user, name, age, _version, _mask)
  select old.rowid
    , case when old.id != new.id then new.id else null end
    , case when old.name != new.name then new.name else null end
    , case when old.age != new.age then new.age else null end
    , new._version
    , (case when old.id != new.id then 1 else 0 end) + (case when old.name != new.name then 2 else 0 end) + (case when old.age != new.age then 4 else 0 end)
  where old.id != new.id or old.name != new.name or old.age != new.age;
end;

