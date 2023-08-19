BEGIN TRANSACTION;
CREATE TABLE IF NOT EXISTS "tags" (
	"id"	integer,
	"created_at"	datetime,
	"updated_at"	datetime,
	"deleted_at"	datetime,
	"name"	text,
	PRIMARY KEY("id")
);
CREATE TABLE IF NOT EXISTS "run_tags" (
	"run_id"	integer,
	"tag_id"	integer,
	PRIMARY KEY("run_id","tag_id"),
	CONSTRAINT "fk_run_tags_tag" FOREIGN KEY("tag_id") REFERENCES "tags"("id"),
	CONSTRAINT "fk_run_tags_run" FOREIGN KEY("run_id") REFERENCES "runs"("id")
);
CREATE TABLE IF NOT EXISTS "runs" (
	"id"	integer,
	"created_at"	datetime,
	"updated_at"	datetime,
	"deleted_at"	datetime,
	"desc"	text,
	"slug"	text UNIQUE,
	"path"	text,
	"remote"	text,
	"finished"	numeric,
	"hidden"	numeric,
	"last_pulled"	datetime,
	"successful"	numeric,
	PRIMARY KEY("id")
);
CREATE INDEX IF NOT EXISTS "idx_tags_deleted_at" ON "tags" (
	"deleted_at"
);
CREATE INDEX IF NOT EXISTS "idx_runs_deleted_at" ON "runs" (
	"deleted_at"
);
CREATE INDEX IF NOT EXISTS "idx_runs_slug" ON "runs" (
	"slug"
);
COMMIT;
