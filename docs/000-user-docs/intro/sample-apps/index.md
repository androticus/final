# Sample Applications

This section walks through **real programs** built in Final — not toy snippets, but small applications that combine language features you have already seen (SQL, tables, linear algebra, formatting) into a coherent story.

| App | Status | Repo path |
| --- | ------ | --------- |
| [Auto-stats](auto-stats.md) | **Done** for Stages 1–3 (wear → class → make) | `internal/final/apps/auto-stats/` |
| Interactive web tutorial | **Planned** | — |

**Auto-stats** asks a practical question: *given a snapshot of used-car dealer listings, how well can we predict asking price from simple vehicle facts?* The narrative builds one regression ladder at a time — mean/age/miles, then vehicle class, then make — and reports honest baselines with PRE/F model comparison. The app sources live under `internal/final/apps/auto-stats/`.

The **interactive web tutorial** will be a separate sample app — a browser-based way to explore Final concepts step by step. Details are still to come; it is listed here so the book has a home for end-to-end examples as they land.

Start with [Auto-stats](auto-stats.md) if you want a data-and-stats story you can run today from the repo.
