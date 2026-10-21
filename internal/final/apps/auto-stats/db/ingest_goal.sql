-- Ingest control / goal tracking for auto-stats (unhuman.autos → unhuman_listing).
-- One row per (API make, API model, year). have_n counts fit-ready rows in
-- unhuman_listing matching db_make/db_model/year (non-null price, miles, year).
-- next_start is the pagination offset for the next fetch (cap ~500 upstream).

CREATE TABLE IF NOT EXISTS ingest_goal (
  make        TEXT    NOT NULL,  -- API query make (e.g. bmw)
  model       TEXT    NOT NULL,  -- API query model (e.g. 3 series)
  year        INT     NOT NULL,  -- model year tranche
  db_make     TEXT    NOT NULL,  -- build.make as stored (e.g. BMW)
  db_model    TEXT    NOT NULL,  -- build.model as stored (e.g. 3 Series)
  zip         TEXT    NOT NULL,
  radius      INT     NOT NULL,
  target_n    INT     NOT NULL,
  have_n      INT     NOT NULL DEFAULT 0,
  next_start  INT     NOT NULL DEFAULT 0,
  PRIMARY KEY (make, model, year)
) STRICT;
