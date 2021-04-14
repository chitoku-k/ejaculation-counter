CREATE TABLE IF NOT EXISTS "counts" (
    "id" SERIAL NOT NULL PRIMARY KEY,
    "user" integer NOT NULL,
    "date" date NOT NULL,
    "count" integer NOT NULL
);
