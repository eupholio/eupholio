# Calculation Methods

## 1) MovingAverage

Maintain state per asset:
- `qty`
- `avg_cost_jpy_per_unit`

### Acquire

`new_avg = (old_qty * old_avg + jpy_cost) / (old_qty + acquire_qty)`

### Dispose

- `cost = sold_qty * avg_cost`
- `realized = jpy_proceeds - cost`
- `qty -= sold_qty`
- `avg_cost` is retained

### Income

Treat as equivalent to `Acquire` and add to inventory (also added to `income_jpy`).

---

## 2) TotalAverage

Aggregate by year and asset:
- `carry_in_qty`, `carry_in_cost`
- `total_acquired_qty`, `total_acquired_cost`
- `total_disposed_qty`, `total_disposed_proceeds`

### Yearly Average

`avg = (carry_in_cost + total_acquired_cost) / (carry_in_qty + total_acquired_qty)`

### Realized Profit/Loss

`realized = total_disposed_proceeds - total_disposed_qty * avg`

### Carry Forward

- `carry_out_qty = (carry_in_qty + total_acquired_qty) - total_disposed_qty`
- `carry_out_cost = carry_out_qty * avg`

Notes:
- `Transfer` is excluded from profit/loss calculation.
- Out-of-year events are excluded with `YearMismatch`.
