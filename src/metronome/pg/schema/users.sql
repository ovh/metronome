CREATE TABLE IF NOT EXISTS users
(
    user_id uuid NOT NULL DEFAULT uuid_generate_v4(),
    name character varying(256) NOT NULL,
    password character varying(256) NOT NULL,
    created_at timestamp without time zone NOT NULL DEFAULT now(),
    roles jsonb,
    CONSTRAINT users_pkey PRIMARY KEY (user_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS users_name_idx
    ON users USING btree
    (name)
    TABLESPACE pg_default;
