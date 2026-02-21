# eupholio-normalizer docs

Documentation for source-specific normalization into `eupholio-core::event::Event`.

- [15-normalizer-bittrex-order-history-mapping.md](./15-normalizer-bittrex-order-history-mapping.md)
  - Concrete phase-1 mapping from Bittrex OrderHistory CSV into Event model
- [16-fixture-policy.md](./16-fixture-policy.md)
  - Fixture naming, anonymization, and regression maintenance rules
- [17-next-adapter-survey.md](./17-next-adapter-survey.md)
  - Candidate selection notes for the next live-source adapter
- [18-go-rust-migration-plan.md](./18-go-rust-migration-plan.md)
  - Staged migration/rollback plan from Go path to Rust path
- [19-normalizer-coincheck-trade-history-mapping.md](./19-normalizer-coincheck-trade-history-mapping.md)
  - Concrete phase-2 mapping for Coincheck trade history CSV (Acquire/Dispose scope)
- [20-normalizer-coincheck-phase3-scope.md](./20-normalizer-coincheck-phase3-scope.md)
  - Scoped plan for adding Received/Sent transfer mapping in Coincheck
- [21-normalizer-coincheck-phase4-scope.md](./21-normalizer-coincheck-phase4-scope.md)
  - Scoped plan for Deposit/Withdrawal (JPY) transfer mapping in Coincheck
- [22-normalizer-bitflyer-phase1-trade-mapping.md](./22-normalizer-bitflyer-phase1-trade-mapping.md)
  - Minimal BitFlyer phase-1 mapping for buy/sell trade rows
