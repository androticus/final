// Cross-platform headless smoke — colon body (if is compile-only).
// Distinctive arith (20+22=42) so a wrong op fails loudly.

: run-headless-smoke
  20 22 + 42 <> if
    "SMOKE_FAIL headless: arith\n" $$
    1 0 /
  then
  "SMOKE_OK headless\n" $
;
run-headless-smoke
