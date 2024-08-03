CREATE TABLE IF NOT EXISTS schema_upgrades (
    version integer NOT NULL,
    date_installed TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW()
);