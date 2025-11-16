-- Bootstrap additional service databases and users for media-api and response-api

-- Create users/roles
DO
$$
BEGIN
	IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'media') THEN
		CREATE ROLE media LOGIN PASSWORD 'media';
	END IF;
END
$$;

DO
$$
BEGIN
	IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'response_api') THEN
		CREATE ROLE response_api LOGIN PASSWORD 'response_api';
	END IF;
END
$$;

-- Create databases (with existence check)
DO
$$
BEGIN
	IF NOT EXISTS (SELECT 1 FROM pg_database WHERE datname = 'media_api') THEN
		CREATE DATABASE media_api;
	END IF;
END
$$;

DO
$$
BEGIN
	IF NOT EXISTS (SELECT 1 FROM pg_database WHERE datname = 'response_api') THEN
		CREATE DATABASE response_api;
	END IF;
END
$$;

-- Grant privileges (these commands are idempotent)
ALTER DATABASE media_api OWNER TO media;
GRANT ALL PRIVILEGES ON DATABASE media_api TO media;

ALTER DATABASE response_api OWNER TO response_api;
GRANT ALL PRIVILEGES ON DATABASE response_api TO response_api;
