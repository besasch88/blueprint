CREATE TABLE "bp_user" (
    "id" varchar(36) PRIMARY KEY NOT NULL,
    "email" varchar(255) NOT NULL,
    "firstname" varchar(255) NOT NULL,
    "lastname" varchar(255) NOT NULL,
    "is_active" boolean NOT NULL,
    "created_at" timestamp NOT NULL,
    "updated_at" timestamp NOT NULL,
    "deleted_at" timestamp
);

ALTER TABLE "bp_user" ADD CONSTRAINT "idx_bp_user_email" UNIQUE ("email");