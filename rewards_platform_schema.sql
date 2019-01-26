/* Rollback tables */
DROP TABLE IF EXISTS public.redemptions_instant;
DROP TABLE IF EXISTS public.redemptions_coupon;
DROP TABLE IF EXISTS public.submissions;
DROP TABLE IF EXISTS public.offers;
DROP TABLE IF EXISTS public.rewards;
DROP TABLE IF EXISTS public.actions;
DROP TABLE IF EXISTS public.stores;
DROP TYPE IF EXISTS "status";

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TYPE "status" AS ENUM ('ACTIVE', 'INACTIVE', 'EXPIRED', 'REDEEMED', 'PENDING', 'ACCEPTED', 'REJECTED');

CREATE TABLE public.stores (
	id SERIAL PRIMARY KEY,
	"name" text NOT NULL,
	postal_code VARCHAR(8) NOT NULL,
	metadata jsonb,
	api_key uuid UNIQUE NOT NULL DEFAULT uuid_generate_v4(),
	"status" status NOT NULL DEFAULT 'ACTIVE',
	created_at timestamptz NOT NULL DEFAULT now(),
	updated_at timestamptz NOT NULL DEFAULT now()
);

INSERT INTO public.stores (id, name, postal_code)
VALUES
 (1,'Tehanos Grill','M1B1K4');

CREATE TABLE public.actions (
	id SERIAL PRIMARY KEY,
	"description" text NOT NULL,
	keywords text,
	created_at timestamptz NOT NULL DEFAULT now(),
	updated_at timestamptz NOT NULL DEFAULT now()
);

INSERT INTO public.actions (id, description)
VALUES
 (1, 'Snap your first bite');

CREATE TABLE public.rewards (
	id SERIAL PRIMARY KEY,
	"description" text NOT NULL,
	keywords text,
	created_at timestamptz NOT NULL DEFAULT now(),
	updated_at timestamptz NOT NULL DEFAULT now()
);

INSERT INTO public.rewards (id, description)
VALUES
 (1, 'Free Dessert'),
 (2, 'Free Fries');


CREATE TABLE public.offers (
	id SERIAL PRIMARY KEY,
	"status" status NOT NULL DEFAULT 'ACTIVE',
	action_id INTEGER REFERENCES actions(id) NOT NULL,
	instant_reward_id INTEGER REFERENCES rewards(id) NOT NULL,
	loyalty_reward_id INTEGER REFERENCES rewards(id) NOT NULL,
	store_id INTEGER REFERENCES stores(id) NOT NULL,
	created_at timestamptz NOT NULL DEFAULT now(),
	updated_at timestamptz NOT NULL DEFAULT now()
);

INSERT INTO public.offers (id, action_id, instant_reward_id, loyalty_reward_id, store_id)
VALUES
 (1, 1, 1, 2, 1);


CREATE TABLE public.submissions (
	id SERIAL PRIMARY KEY,
	instagram_account VARCHAR(35) NOT NULL,
	follower_count INTEGER NOT NULL,
	metadata jsonb DEFAULT '[{}]',
	"status" status NOT NULL DEFAULT 'PENDING',
	offer_id INTEGER REFERENCES offers(id) NOT NULL,
	instant_reward_expire_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP + interval '2 day',
	created_at timestamptz NOT NULL DEFAULT now(),
	updated_at timestamptz NOT NULL DEFAULT now()
);

INSERT INTO public.submissions (id, instagram_account, follower_count, offer_id)
VALUES
 (1, '@ahmed.dauda', 210, 1),
 (2, '@lolade.ship', 300, 1);

CREATE TABLE public.redemptions_instant (
	id SERIAL PRIMARY KEY,
	submission_id INTEGER REFERENCES submissions(id) NOT NULL,
	redeemed_at timestamptz
);

INSERT INTO public.redemptions_instant (submission_id)
VALUES
 (1);

CREATE TABLE public.redemptions_coupon (
	id SERIAL PRIMARY KEY,
	code VARCHAR(5) NOT NULL DEFAULT UPPER(LEFT(MD5(random()::text),4)),
	"status" status NOT NULL DEFAULT 'PENDING',
	submission_id INTEGER REFERENCES submissions(id) NOT NULL,
	expire_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP + interval '3 months',
	created_at timestamptz NOT NULL DEFAULT now(),
	redeemed_at timestamptz
);

INSERT INTO public.redemptions_coupon (id, submission_id)
VALUES
 (1, 2);