CREATE TABLE IF NOT EXISTS tasks
(
    guid text NOT NULL,
    user_id uuid NOT NULL,
    name text NOT NULL,
    urn text NOT NULL,
    schedule text NOT NULL,
    created_at timestamp without time zone NOT NULL,
    id text NOT NULL,
    CONSTRAINT tasks_pkey PRIMARY KEY (guid),
    CONSTRAINT user_id_fk FOREIGN KEY (user_id)
        REFERENCES users (user_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
)
