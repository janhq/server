-- Bootstrap additional service databases and users for media-api and response-api

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


CREATE DATABASE media_api OWNER media;
GRANT ALL PRIVILEGES ON DATABASE media_api TO media;

CREATE DATABASE response_api OWNER response_api;
GRANT ALL PRIVILEGES ON DATABASE response_api TO response_api;
