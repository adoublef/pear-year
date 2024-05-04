create table users (
  -- format as uuid-v7
  id text
  , name text not null
  , dob real not null
  , role text not null
  , _version int not null
  , check (length(name) <= 30)
  , check (role in ('Guest', 'Support', 'Admin'))
  , primary key (id)
) strict;

create table _users_history (
  _rowid int
  , user text
  , name text
  , dob real
  , role text
  , _version int not null
  , _mask int not null
  , primary key (_rowid, _version)
) without rowid;

create trigger users_update_history
after update on users for each row
begin
  insert into _users_history (_rowid, user, name, dob, role, _version, _mask)
  select old.rowid
    , case when old.id != new.id then old.id else null end
    , case when old.name != new.name then old.name else null end
    , case when old.dob != new.dob then old.dob else null end
    , case when old.role != new.role then old.role else null end
    , new._version - 1
    , ((old.id != new.id) << 0) 
    | ((old.name != new.name) << 1) 
    | ((old.dob != new.dob) << 2)
    | ((old.role != new.role) << 3)
  where old.id != new.id or old.name != new.name or old.dob != new.dob or old.role != new.role;
end;

