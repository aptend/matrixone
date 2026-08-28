# TPC-DS 1TB Q11+ plan and execution baseline

This directory is the append-as-diagnosed baseline for queries after Q10. Every
query is recorded immediately after its 1TB execution succeeds or after its plan
is deliberately stopped as unsafe. Ordinary, non-verbose `EXPLAIN` is a
structural diagnostic; only a successful SQL result is a black-box execution
oracle.

## Baseline identity

- Database: `tpcds_1t`
- MatrixOne version string: `8.0.30-MatrixOne-v2`
- Git HEAD: `53ff1f5b8523768da9bf45816cadc7fb71d6146a`
- `mo-service` SHA-256: `d0d0ccd30f86d4072c4ed7af3f09858adb29d1190be6a42a2012d7586e0b31ee`
- Tracked `pkg/sql/plan` production diff SHA-256: `fd98a768ca036941dff3be470b65ef249210e69b2097d6508940291a7aaf2601`
- Auto merge disabled; spill thresholds are 128 MiB.

## Query status

| Query | SQL | Ordinary EXPLAIN | Status | Statement ID | Time | Rows |
|---|---|---|---|---|---:|---:|
| Q11 | [`q11.sql`](queries/q11.sql) | [`q11.explain.txt`](explain/q11.explain.txt) | Success | `01a041e5-aeff-7478-b595-7fe670a3ebf4` | 254.265 s | 100 |
| Q12 | [`q12.sql`](queries/q12.sql) | [`q12.explain.txt`](explain/q12.explain.txt) | Success | `01a041e9-f9a7-7dfd-856e-e26d64749231` | 10.827 s | 100 |
| Q13 | [`q13.sql`](queries/q13.sql) | [`q13.explain.txt`](explain/q13.explain.txt) | Success | `01a041ea-88d1-7c5b-9a98-be5a864cbb71` | 169.422 s | 1 |
| Q14a | [`q14a.sql`](queries/q14a.sql) | [`q14a.explain.txt`](explain/q14a.explain.txt) | Not run: unsafe scan amplification | — | — | — |
| Q14b | [`q14b.sql`](queries/q14b.sql) | [`q14b.explain.txt`](explain/q14b.explain.txt) | Not run: diagnosed with Q14a | — | — | — |

`stats/` contains a `table_stats()` snapshot for every table referenced by the
currently recorded queries. `SHA256SUMS` covers all SQL, EXPLAIN, and stats files.
