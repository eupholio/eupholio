# eupholio-core docs

This directory contains design and operational documentation for `eupholio-core` (the Rust calculation core).

## Table of contents

- [01-overview.md](./01-overview.md)
  - Purpose, scope, and out-of-scope items
- [02-domain-model.md](./02-domain-model.md)
  - Event/Config/Report specifications
- [03-calculation-methods.md](./03-calculation-methods.md)
  - Detailed calculations for MovingAverage / TotalAverage
- [04-cli.md](./04-cli.md)
  - JSON I/O and execution examples for the CLI
- [05-parity-testing.md](./05-parity-testing.md)
  - Parity validation workflow against the Go implementation
- [06-roadmap.md](./06-roadmap.md)
  - Planned future extensions
- [07-rounding-policy.md](./07-rounding-policy.md)
  - Policy for externally injected rounding rules
- [08-normalizer-interface.md](./08-normalizer-interface.md)
  - Draft interface for normalizing exchange inputs into Events
- [09-validation-codes.md](./09-validation-codes.md)
  - List of issue codes returned by `validate`
- [10-pr-prep-core-rs.md](./10-pr-prep-core-rs.md)
  - PR preparation notes for merging core-rs into main
- [11-pr-ready-summary-core-rs.md](./11-pr-ready-summary-core-rs.md)
  - Collected summary for PR readiness checkpoints
- [12-dev-workflows.md](./12-dev-workflows.md)
  - Reusable development and review loops
