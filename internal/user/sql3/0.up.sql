create table users (
  -- format as uuid-v7
  id text,
  name text not null,
  dob real not null,
  role text not null,
  _version int not null,
  check (length(name) <= 30)
  check (role in ('Guest', 'Support', 'Admin'))
  primary key (id)
) strict;

create table _users_history (
  _rowid int,
  user text,
  name text,
  dob real,
  role text,
  _version int not null,
  -- mask (https://simonwillison.net/2023/Apr/15/sqlite-history/)
  _mask int not null,
  check (_version >= 1)
  primary key (_rowid, _version)
) without rowid;

create trigger users_insert_history 
after insert on users
begin
  insert into _users_history (_rowid, user, name, dob, role, _version, _mask)
  values (
    new.rowid,
    new.id,
    new.name,
    new.dob,
    new.role,
    1,
    (1 << 4) - 1
  );
end;

create trigger users_update_history
after update on users for each row
begin
  insert into _users_history (_rowid, user, name, dob, role, _version, _mask)
  select old.rowid
    , case when old.id != new.id then new.id else null end
    , case when old.name != new.name then new.name else null end
    , case when old.dob != new.dob then new.dob else null end
    , case when old.role != new.role then new.role else null end
    , new._version
    , ((old.id != new.id) << 0) 
    | ((old.name != new.name) << 1) 
    | ((old.dob != new.dob) << 2)
    | ((old.role != new.role) << 3)
  where old.id != new.id or old.name != new.name or old.dob != new.dob or old.role != new.role;
end;

