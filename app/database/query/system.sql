-- name: GetCurrentDatabase :one
SELECT current_database() as database_name;

-- name: GetDatabaseVersion :one
SELECT version() as version;

-- name: GetConnectionInfo :one
SELECT 
    current_database() as database_name,
    current_user as username,
    inet_client_addr() as client_address,
    inet_client_port() as client_port;

-- -- name: CheckDatabaseExists :one
-- SELECT EXISTS(
--     SELECT 1 FROM pg_database WHERE datname = $1
-- ) as exists;

-- -- name: GetActiveConnections :one
-- SELECT count(*) as active_connections 
-- FROM pg_stat_activity 
-- WHERE datname = current_database();