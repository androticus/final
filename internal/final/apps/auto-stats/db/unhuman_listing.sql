-- Proposed raw listing subset for auto-stats (unhuman.autos / MarketCheck proxy).
-- Source shape: top-level id/price/miles; vehicle fields under nested `build`
-- (llms.txt flat schema is inaccurate — use sample unhuman-auto.json).
--
-- Ingest: project/flatten only these columns; drop media[], dealer{}, mc_dealership{},
-- and other noise. Do not store derived ageYears here — compute at read/fit time from year.
--
-- Used only: API queries set inventory_type=used; rows with inventory_type other than
-- used (or near-zero miles) are rejected at ingest. See 160.2.
--
-- STRICT aligns with Final Phase 0 SQLite (docs/003-design-notes/152.2-db-sqlite.md).
-- Payload columns nullable (incomplete listings / missing price). No secondary indexes:
-- browse/summarize/DISTINCT on a small in-memory-friendly extract.

CREATE TABLE IF NOT EXISTS unhuman_listing (
  id              TEXT    PRIMARY KEY,  -- listing id (upsert key)
  price           INT,                  -- asking USD (top-level); NULL if absent
  miles           INT,                  -- odometer (top-level)
  year            INT,                  -- build.year (model year; age = as_of - year)
  make            TEXT,                 -- build.make
  model           TEXT,                 -- build.model
  body_type       TEXT,                 -- build.body_type  (e.g. Sedan, SUV)
  vehicle_type    TEXT,                 -- build.vehicle_type (e.g. Car, Truck)
  drivetrain      TEXT,                 -- build.drivetrain
  doors           INT,                  -- build.doors
  cylinders       INT,                  -- build.cylinders
  inventory_type  TEXT                  -- used | new | cpo (ingest keeps used only)
) STRICT;

-- Upsert sketch (host / Final later):
--   INSERT INTO unhuman_listing (…) VALUES (…)
--   ON CONFLICT(id) DO UPDATE SET
--     price=excluded.price,
--     miles=excluded.miles,
--     year=excluded.year,
--     make=excluded.make,
--     model=excluded.model,
--     body_type=excluded.body_type,
--     vehicle_type=excluded.vehicle_type,
--     drivetrain=excluded.drivetrain,
--     doors=excluded.doors,
--     cylinders=excluded.cylinders,
--     inventory_type=excluded.inventory_type;
