CREATE TABLE IF NOT EXISTS "counts" (
    "user_id" integer NOT NULL,
    "date" date NOT NULL,
    "count" integer NOT NULL,
    UNIQUE ("user_id", "date")
);
