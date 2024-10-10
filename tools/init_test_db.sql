-- Make sure to connect to the test_db before running these commands
-- otherwise the public schema will be of the databse your connected to
-- and not the 'test_db' you created above.
CREATE DATABASE test_db;

BEGIN;
CREATE ROLE test_db_role;
GRANT ALL PRIVILEGES ON DATABASE test_db TO test_db_role;
GRANT ALL PRIVILEGES ON SCHEMA public TO test_db_role;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO test_db_role;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO test_db_role;
GRANT ALL PRIVILEGES ON ALL FUNCTIONS IN SCHEMA public TO test_db_role;
ALTER DEFAULT PRIVILEGES for role postgres IN SCHEMA public 
  GRANT ALL PRIVILEGES ON TABLES TO test_db_role;
ALTER DEFAULT PRIVILEGES for role postgres IN SCHEMA public 
  GRANT ALL PRIVILEGES ON SEQUENCES TO test_db_role;
ALTER DEFAULT PRIVILEGES for role postgres IN SCHEMA public
  GRANT ALL PRIVILEGES ON FUNCTIONS TO test_db_role;

CREATE USER test_user WITH ENCRYPTED PASSWORD 'test_user_password' LOGIN;
GRANT test_db_role TO test_user;
COMMIT;

-- BEGIN;
-- REVOKE test_db_role FROM test_user;
-- DROP USER IF EXISTS test_user;
-- ALTER DEFAULT PRIVILEGES for role postgres IN SCHEMA public REVOKE ALL PRIVILEGES ON FUNCTIONS FROM test_db_role;
-- ALTER DEFAULT PRIVILEGES for role postgres IN SCHEMA public REVOKE ALL PRIVILEGES ON SEQUENCES FROM test_db_role;
-- ALTER DEFAULT PRIVILEGES for ROLE postgres IN SCHEMA public REVOKE ALL PRIVILEGES ON TABLES FROM test_db_role;
-- REVOKE ALL PRIVILEGES ON ALL FUNCTIONS IN SCHEMA public FROM test_db_role;
-- REVOKE ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public FROM test_db_role;
-- REVOKE ALL PRIVILEGES ON ALL TABLES IN SCHEMA public FROM test_db_role;
-- REVOKE ALL PRIVILEGES ON SCHEMA public FROM test_db_role;
-- REVOKE ALL PRIVILEGES ON DATABASE test_db FROM test_db_role;
-- DROP ROLE IF EXISTS test_db_role;
-- COMMIT;

-- DROP DATABASE test_db;