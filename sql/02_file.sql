DROP TABLE IF EXISTS files;

CREATE TABLE files(
  id  serial PRIMARY_KEY,
  user_id reference users(id) not null default 1 ON DELETE CASCADE, 
  name varchar(250) not null default '',
);