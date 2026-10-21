# Auto-stats

**Auto-stats** predicts used-car **asking price** from listing data. In the repo the app lives under `internal/final/apps/auto-stats/`. Data comes from a frozen SQLite extract of dealer listings sourced from [unhuman.autos](https://unhuman.autos), stored as `internal/final/apps/auto-stats/db/auto-stats.db`.

This chapter walks three OLS stages on that freeze:

1. **Stage 1** — mean baseline, then age and miles (alone, then together)
2. **Stage 2** — vehicle class (car / SUV / truck) on top of age and miles
3. **Stage 3** — make (brand), still with age and miles, but **without** vehicle class

---

## The question and the data

We want a number: *what should this listing cost?* A naive first answer is the **average price in the dataset** — anything fancier has to beat that on error (root-mean-square error, **RMSE**). We elaborate the model in stages, hoping each step predicts asking price better from facts such as age, mileage, class, and make. Alongside RMSE we report **R²** (correlation-style fit): **0** means no predictive value beyond the mean; **1** would be perfect prediction (rare in real markets). The goal at each step is **less error** and a **higher R²**.

Fit-ready rows for this chapter:

| Filter | Value |
| ------ | ----- |
| Listings | **n ≈ 2817** (2811 after class mapping) |
| As-of year | **2026** |
| Condition | used only (`inventory_type` null or `'used'`) |
| Miles | ≥ **500** (drop near-new placeholders) |
| Outcome **y** | asking price (USD) |

Age and miles **covary**: older cars tend to have more miles. Stage 1 therefore fits them **alone first**, then **together**, and compares nested models rather than jumping to a kitchen-sink equation. Later stages use the same compact-vs-augmented (**C-vs-A**) habit: keep a simpler model **C**, add a block in **A**, and ask how much of C’s remaining error A removes.

**PRE** (*proportional reduction in error*) is that fraction: \((\mathrm{SSE}_C - \mathrm{SSE}_A) / \mathrm{SSE}_C\). An accompanying **F** statistic asks whether the PRE is large relative to the extra parameters and leftover noise.

---

## Stage 1 — mean, age, and miles

### Mean-only baseline

If every car were priced at the sample mean, **R² ≈ 0** by definition and **RMSE ≈ $12,816**. The intercept is about **$29,008** — the average asking price in this slice of the market. That wide scatter is the benchmark every richer model must improve on.

### Age alone

One extra parameter (age in years) lifts **R² to 0.23** and cuts **RMSE to ~$11,251**. The slope is about **−$3,075 per year**: holding nothing else fixed, each additional year of age is associated with roughly three thousand dollars less on the sticker.

Compared to the mean-only model: **PRE ≈ 0.23**, **F ≈ 839** — age explains about 23% of what the mean left on the table.

### Miles alone

Miles alone does better as a **single** predictor: **R² = 0.33**, **RMSE ≈ $10,529**, slope about **−$0.19 per mile**. vs mean: **PRE ≈ 0.33**, **F ≈ 1357**. **Better single = miles** — odometer reading alone beats age alone on this dataset.

### Age and miles together

Adding both predictors reaches **R² = 0.35** and **RMSE ≈ $10,342** (three parameters: intercept, age, miles). Coefficients are roughly **$43,289** intercept, **−$1,279/year**, **−$0.15/mile**. vs mean: **PRE ≈ 0.35**, **F ≈ 755** (two added df).

Incremental value after each single predictor:

| Comparison | PRE | F |
| ---------- | --- | - |
| mean → age+miles | ≈ 0.35 | ≈ 755 |
| age → age+miles | ≈ 0.16 | ≈ 517 |
| miles → age+miles | ≈ 0.036 | ≈ 104 |

**Age still adds clear incremental PRE after miles** — the two signals are related but not redundant. Even so, the joint model only nibbles another ~$200 off RMSE compared to miles alone; most of the Stage-1 gain comes from getting odometer on the table.

### What Stage 1 leaves on the floor

Results **do** beat the mean, but **residual RMSE is still ~$10k**. Picture two listings with the same age and miles: a base Honda Civic and a lightly used luxury sedan can sit worlds apart in price. Age and miles explain broad depreciation; they do not know *what* the vehicle is.

That gap is the bridge to Stage 2.

Run: `internal/final/apps/auto-stats/ladder_age_miles.fnl` → word `ladder`.

---

## Stage 2 — vehicle class (car / SUV / truck)

Age and miles still leave ~$10k of RMSE. Intuition says **what kind of vehicle** matters: trucks tend to cost more than SUVs, which tend to cost more than passenger cars — even at similar age and miles. Stage 2 elaborates the Stage-1 wear core with a **vehicle-class** block.

### Coding note

The listing feed’s raw `vehicle_type` field is only **Car** or **Truck**, and **SUVs are lumped under Truck**. That collapses the story we care about. For this ladder we derive class from `body_type` instead:

| Derived class | `body_type` values |
| ------------- | ------------------ |
| **car** | Sedan, Coupe, Hatchback |
| **suv** | SUV, Wagon (e.g. Outback) |
| **truck** | Pickup |

Unmapped body types are dropped (six rows here). Fit **n = 2811**. **Car** is the reference category; the design matrix adds **isSUV** and **isTruck** (with intercept). Compact **C** = age + miles; augmented **A** = age + miles + class.

Raw class means in this extract already line up with the narrative:

| Class | n | Mean asking price |
| ----- | - | ----------------- |
| car | 1078 | ≈ $24,772 |
| suv | 1437 | ≈ $30,517 |
| truck | 296 | ≈ $37,370 |

### Results — and how they match intuition

| Model | p | R² | RMSE |
| ----- | - | -- | ---- |
| age+miles | 3 | ≈ 0.35 | ≈ $10,344 |
| age+miles+class | 5 | ≈ 0.43 | ≈ $9,660 |

**C-vs-A** (wear → wear+class): **PRE ≈ 0.13**, **F ≈ 207** (df1 = 2). Class eats about **13%** of the error that age and miles left behind — a clear jump, not a rounding error. RMSE drops by roughly **$700**.

Relative to a **car** of the same age and miles, the fitted level shifts are about **+$4,600 for an SUV** and **+$12,500 for a truck** (with age ≈ −$1,285/yr and miles ≈ −$0.15/mi still in the model).

That is exactly the ordering we expected going in:

1. **Raw means** already ranked truck ≫ SUV ≫ car (~$37k / $31k / $25k).
2. **After controlling for age and miles**, the same ranking survives in the βs — so this is not just “trucks are newer” or “SUVs have fewer miles.” Class carries a **price-level** story of its own.
3. The **gap sizes** feel right too: SUV is a moderate step above car; full-size pickup is a much larger step — which matches how people shop those segments.

So the data-driven ladder does not overturn the garage-floor intuition; it **quantifies** it (PRE, dollars of shift, remaining RMSE) and confirms the shifts are still there once wear is in the model.

### What Stage 2 leaves on the floor

Vehicle class helps — but **RMSE is still ~$9.7k**. A nicely equipped **BMW 3 Series** can still ask more than a typical **RAV4** or **CR-V** at similar wear. Class is a coarse bucket; **brand and nameplate** sit inside those buckets. That motivates Stage 3.

Run: `internal/final/apps/auto-stats/ladder_vehicle_type.fnl` → word `ladderVt`.

---

## Stage 3 — make (brand), without vehicle class

Luxury-vs-crossover prices show why class alone will not finish the job. Body class also **correlates strongly with make and model mix** (Ford’s F-150 vs Honda’s CR-V; BMW’s 3 Series + X5 under one brand flag). So Stage 3 **drops vehicle class** from the design matrix and asks a different question: given age and miles, how much does **make** buy?

**Honda** is the reference make (a mass-market mid). Other catalog makes get one-hot columns. Compact **C** = age + miles; augmented **A** = age + miles + make. Fit **n = 2817**.

### Results

| Model | p | R² | RMSE |
| ----- | - | -- | ---- |
| age+miles | 3 | ≈ 0.35 | ≈ $10,342 |
| age+miles+make | 14 | ≈ **0.61** | ≈ **$8,013** |

**C-vs-A** (wear → wear+make): **PRE ≈ 0.40**, **F ≈ 171** (df1 = 11). Make removes about **40%** of the error left by age and miles — a much larger bite than class’s ~13%. R² jumps from the mid‑0.3s into the **0.60s**; RMSE falls from ~$10.3k into the **~$8k** band.

Wear coefficients remain sensible with make in the model: about **−$1,723 per year** and **−$0.11 per mile**.

Make level shifts vs **Honda** (same age and miles), rounded:

| Make | Δ asking $ | Make | Δ asking $ |
| ---- | ---------- | ---- | ---------- |
| RAM | ≈ +$21,800 | Lincoln | ≈ +$17,500 |
| Lexus | ≈ +$14,000 | BMW | ≈ +$9,700 |
| Ford | ≈ +$9,100 | Dodge | ≈ +$7,200 |
| Toyota | ≈ +$2,100 | Audi | ≈ +$1,100 |
| Subaru | ≈ $0 | Chevrolet | ≈ −$400 |
| **Honda** | reference | Nissan | ≈ −$5,600 |

### How this meshes with intuition

- **Luxury and truck-heavy brands** sit well above Honda (Lincoln, Lexus, BMW, RAM, Ford-with-F-150 mix) — the BMW‑vs‑RAV4 story shows up as a brand-level premium even without a separate class column.
- **Near-Honda mass brands** (Toyota, Subaru, Chevrolet, Audi in this catalog mix) cluster near the reference.
- **Nissan** lands clearly cheaper at the same age and miles in this freeze — a discount the model is allowed to report without us forcing it.

One caveat, by design: a make dummy absorbs that brand’s **model mix** (BMW = sedan + SUV; Ford = pickup + Explorer). That is why we removed class first — class and make are entangled, and Stage 3 asks for brand identity’s predictive power on top of wear, not a pure “logo only” effect with class held fixed.

### Where the ladder stands

| Stage | Story | R² (best in stage) | RMSE (best) |
| ----- | ----- | ------------------ | ----------- |
| 1 | mean → age / miles | ≈ 0.35 | ≈ $10.3k |
| 2 | + vehicle class | ≈ 0.43 | ≈ $9.7k |
| 3 | + make (no class) | ≈ **0.61** | ≈ **$8.0k** |

From “guess the average” (~$12.8k RMSE) to age+miles+make (~$8.0k), the ladder cuts a large share of error and more than doubles R² relative to wear alone — while still leaving a several-thousand-dollar cloud. Trim, options, condition, and exact nameplate are among the reasons; those are natural later elaborations if the catalog and *n* support them.

Run: `internal/final/apps/auto-stats/ladder_make.fnl` → word `ladderMake`.

---

## Run it yourself

The three stage scripts live under `internal/final/apps/auto-stats/`.

From the **repository root**:

```forth
bin/final-go -core
"internal/final/apps/auto-stats/ladder_age_miles.fnl" compileFromFile
ladder

"internal/final/apps/auto-stats/ladder_vehicle_type.fnl" compileFromFile
ladderVt

"internal/final/apps/auto-stats/ladder_make.fnl" compileFromFile
ladderMake
```

Each script opens the frozen SQLite DB, builds design matrices, fits OLS models, and prints ladder lines plus C-vs-A PRE/F. Ingest and schema live alongside them if you want to trace data from JSON export to query.

Related language topics: [SQL Databases](../database.md) for `DbProvider` and queries, [Table](../compounds/table.md) for row-shaped results, and [Linear Algebra](../la/index.md) (`VecNf`, `Mat2f`, `OLSModel`).
