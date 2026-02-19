# Calculation Methods

## 1) MovingAverage

資産ごとに状態を保持:
- `qty`
- `avg_cost_jpy_per_unit`

### Acquire

`new_avg = (old_qty * old_avg + jpy_cost) / (old_qty + acquire_qty)`

### Dispose

- `cost = sold_qty * avg_cost`
- `realized = jpy_proceeds - cost`
- `qty -= sold_qty`
- `avg_cost` は維持

### Income

Acquire相当として在庫へ取り込む（`income_jpy` にも加算）

---

## 2) TotalAverage

年次・資産ごとに集計:
- `carry_in_qty`, `carry_in_cost`
- `total_acquired_qty`, `total_acquired_cost`
- `total_disposed_qty`, `total_disposed_proceeds`

### 年次平均

`avg = (carry_in_cost + total_acquired_cost) / (carry_in_qty + total_acquired_qty)`

### 実現損益

`realized = total_disposed_proceeds - total_disposed_qty * avg`

### 繰越

- `carry_out_qty = (carry_in_qty + total_acquired_qty) - total_disposed_qty`
- `carry_out_cost = carry_out_qty * avg`

補足:
- `Transfer` は損益計算対象外
- 年次外イベントは `YearMismatch` で除外
