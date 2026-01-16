CREATE TYPE "AccountStatus" AS ENUM (
  'active',
  'frozen',
  'closed'
);

CREATE TYPE "TransactionType" AS ENUM (
  'deposit',
  'withdrawal',
  'transfer_in',
  'transfer_out'
);

CREATE TYPE "TransferStatus" AS ENUM (
  'pending',
  'completed',
  'failed',
  'cancelled',
  'reversed'
);

CREATE TYPE "Currency" AS ENUM (
  'USD',
  'EUR',
  'GBP',
  'BDT',
  'INR'
);

CREATE TABLE "users" (
  "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
  "first_name" varchar NOT NULL,
  "last_name" varchar NOT NULL,
  "email" varchar UNIQUE NOT NULL,
  "password_hash" varchar NOT NULL,
  "created_at" timestamptz DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "accounts" (
  "id" bigserial PRIMARY KEY,
  "owner_id" uuid NOT NULL,
  "balance_cents" bigint NOT NULL DEFAULT 0,
  "currency" "Currency" NOT NULL,
  "status" "AccountStatus" NOT NULL DEFAULT 'active',
  "created_at" timestamptz DEFAULT (now())
);

CREATE TABLE "transactions" (
  "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
  "account_id" bigint NOT NULL,
  "type" "TransactionType" NOT NULL,
  "amount_cents" bigint NOT NULL,
  "balance_after_cents" bigint NOT NULL,
  "related_account_id" bigint,
  "reference" varchar UNIQUE,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "transfers" (
  "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
  "from_account_id" bigint NOT NULL,
  "to_account_id" bigint NOT NULL,
  "amount_cents" bigint NOT NULL,
  "status" "TransferStatus" NOT NULL DEFAULT 'pending',
  "created_at" timestamptz DEFAULT (now()),
  "processed_at" timestamptz
);

CREATE INDEX ON "accounts" ("owner_id");

CREATE INDEX ON "accounts" ("owner_id", "status");

CREATE INDEX ON "transactions" ("account_id");

CREATE INDEX ON "transactions" ("account_id", "created_at");

CREATE INDEX ON "transfers" ("from_account_id");

CREATE INDEX ON "transfers" ("to_account_id");

CREATE INDEX ON "transfers" ("from_account_id", "to_account_id");

COMMENT ON TABLE "users" IS 'Stores registered users of the system.';

COMMENT ON TABLE "accounts" IS 'Each bank account belongs to a user and holds a balance in a fixed currency.';

COMMENT ON COLUMN "accounts"."balance_cents" IS 'Balance stored in cents; never negative unless overdraft allowed';

COMMENT ON TABLE "transactions" IS 'Immutable ledger of all money movements. One record per event.';

COMMENT ON COLUMN "transactions"."related_account_id" IS 'Used for transfers';

COMMENT ON COLUMN "transactions"."reference" IS 'Idempotency key / external reference';

COMMENT ON TABLE "transfers" IS 'Represents an intention to move money. A transfer usually results in two transaction records (out + in).';

COMMENT ON COLUMN "transfers"."amount_cents" IS 'Always positive; transfer direction determined by from/to accounts';

COMMENT ON COLUMN "transfers"."processed_at" IS 'Populated when completed or failed';

ALTER TABLE "accounts" ADD FOREIGN KEY ("owner_id") REFERENCES "users" ("id");

ALTER TABLE "transactions" ADD FOREIGN KEY ("account_id") REFERENCES "accounts" ("id");

ALTER TABLE "transactions" ADD FOREIGN KEY ("related_account_id") REFERENCES "accounts" ("id");

ALTER TABLE "transfers" ADD FOREIGN KEY ("from_account_id") REFERENCES "accounts" ("id");

ALTER TABLE "transfers" ADD FOREIGN KEY ("to_account_id") REFERENCES "accounts" ("id");
