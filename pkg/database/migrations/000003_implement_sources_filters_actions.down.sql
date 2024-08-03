DROP TABLE if exists filter;

DROP TABLE IF EXISTS rule;

DROP TABLE IF EXISTS action;

DROP TABLE IF EXISTS profile_sources;

DROP TABLE IF EXISTS source;

CREATE TABLE IF NOT EXISTS module (
    id         integer PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    profile    text references profile(name) NOT NULL,
    name       text NOT NULL,
    parameters jsonb NOT NULL
);

ALTER TABLE profile ADD COLUMN source text NOT NULL;