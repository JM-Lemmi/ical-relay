CREATE TABLE IF NOT EXISTS notifier_source (
    id  integer PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    url text NOT NULL
);

CREATE TABLE IF NOT EXISTS notifier_email_subscriber (
    id  integer PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    source integer REFERENCES notifier_source(id) ON DELETE CASCADE NOT NULL,
    subscribed_at timestamp with time zone default now(), 
    email text NOT NULL,
    token text NOT NULL, /* token for subscribing and unsubscribing */
    validated bool NOT NULL DEFAULT FALSE,
    last_notified timestamp with time zone,
);

CREATE TABLE IF NOT EXISTS notifier_rss_output (
    id  integer PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    source integer REFERENCES notifier_source(id) ON DELETE CASCADE NOT NULL,
    file_path text NOT NULL,
);

CREATE TABLE IF NOT EXISTS notifier_history (
    id integer PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    source integer REFERENCES notifier_source(id) ON DELETE CASCADE NOT NULL,
)
