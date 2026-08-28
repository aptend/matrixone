# TPC-DS 1TB Q1–Q10 plan baseline

This directory records the ordinary, non-verbose `EXPLAIN` output used after the
Q1–Q10 investigation. It is a structural regression baseline, not a query-result
correctness oracle.

## Baseline identity

- Captured: `2026-08-27 14:21:35 CST`
- Database: `tpcds_1t`
- SQL account/user: `sys` / `dump`
- MatrixOne version string: `8.0.30-MatrixOne-v2`
- Git HEAD: `53ff1f5b8523768da9bf45816cadc7fb71d6146a`
- `mo-service` SHA-256: `d0d0ccd30f86d4072c4ed7af3f09858adb29d1190be6a42a2012d7586e0b31ee`
- Tracked `pkg/sql/plan` production diff SHA-256: `fd98a768ca036941dff3be470b65ef249210e69b2097d6508940291a7aaf2601`
- Auto merge: disabled; merge queue length: `0`
- `agg_spill_mem`, `join_spill_mem`, `sort_spill_mem`: `134217728` bytes each

The binary was built from the dirty working tree identified above, including the
Q2 CTE reuse, Q4 partial aggregate, and Q10 existential MARK-join changes. The
production-diff hash excludes test and documentation files so later snapshots can
identify the planner implementation independently of this baseline directory.

## Layout

- `queries/qNN.sql`: exact parameterized SQL used to produce each plan.
- `explain/qNN.explain.txt`: raw ordinary `EXPLAIN` output, without `VERBOSE`,
  table formatting, or post-processing.
- `stats/*.table_stats.txt`: `table_stats()` snapshot for every table scanned by
  Q1–Q10. This distinguishes optimizer changes from changes in the input stats.
- `SHA256SUMS`: integrity hashes for all recorded SQL, plans, and stats.

| Query | SQL | Ordinary EXPLAIN |
|---|---|---|
| Q1 | [`q01.sql`](queries/q01.sql) | [`q01.explain.txt`](explain/q01.explain.txt) |
| Q2 | [`q02.sql`](queries/q02.sql) | [`q02.explain.txt`](explain/q02.explain.txt) |
| Q3 | [`q03.sql`](queries/q03.sql) | [`q03.explain.txt`](explain/q03.explain.txt) |
| Q4 | [`q04.sql`](queries/q04.sql) | [`q04.explain.txt`](explain/q04.explain.txt) |
| Q5 | [`q05.sql`](queries/q05.sql) | [`q05.explain.txt`](explain/q05.explain.txt) |
| Q6 | [`q06.sql`](queries/q06.sql) | [`q06.explain.txt`](explain/q06.explain.txt) |
| Q7 | [`q07.sql`](queries/q07.sql) | [`q07.explain.txt`](explain/q07.explain.txt) |
| Q8 | [`q08.sql`](queries/q08.sql) | [`q08.explain.txt`](explain/q08.explain.txt) |
| Q9 | [`q09.sql`](queries/q09.sql) | [`q09.explain.txt`](explain/q09.explain.txt) |
| Q10 | [`q10.sql`](queries/q10.sql) | [`q10.explain.txt`](explain/q10.explain.txt) |

## Regeneration and comparison

Generate a candidate plan with the same SQL and database:

```sh
{ printf 'EXPLAIN '; sed -n '1,$p' queries/q01.sql; } |
  mysql -h127.0.0.1 -P6001 -udump tpcds_1t \
    --batch --raw --skip-column-names > /tmp/q01.explain.txt
diff -u explain/q01.explain.txt /tmp/q01.explain.txt
```

Interpret a text difference together with the saved stats and semantic regression
tests. A changed node order or rendering is not by itself a correctness regression;
query results remain the black-box oracle.
