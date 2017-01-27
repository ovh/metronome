CREATE TABLE IF NOT EXISTS tokens
(
    token text NOT NULL,
    user_id uuid NOT NULL,
    type character varying(256) NOT NULL,
    created_at timestamp without time zone NOT NULL DEFAULT now(),
    roles jsonb,
    CONSTRAINT tokens_pkey PRIMARY KEY (token),
    CONSTRAINT user_id_fk FOREIGN KEY (user_id)
        REFERENCES users (user_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);
CREATE UNIQUE INDEX IF NOT EXISTS tokens_token_idx
    ON tokens USING btree
    (token)
    TABLESPACE pg_default;
