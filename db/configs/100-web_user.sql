CREATE ROLE akdb_role;
GRANT CONNECT ON DATABASE akdb TO akdb_role;
GRANT USAGE ON SCHEMA public TO akdb_role;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO akdb_role;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO akdb_role;
GRANT ALL PRIVILEGES ON ALL FUNCTIONS IN SCHEMA public TO akdb_role;
ALTER DEFAULT PRIVILEGES for role postgres IN SCHEMA public 
  GRANT INSERT, UPDATE, DELETE, SELECT ON TABLES TO akdb_role;
ALTER DEFAULT PRIVILEGES for role postgres IN SCHEMA public 
  GRANT ALL PRIVILEGES ON SEQUENCES TO akdb_role;
ALTER DEFAULT PRIVILEGES for role postgres IN SCHEMA public
  GRANT ALL PRIVILEGES ON FUNCTIONS TO akdb_role;

CREATE USER web_user WITH ENCRYPTED PASSWORD 'password' LOGIN;
GRANT akdb_role TO web_user;