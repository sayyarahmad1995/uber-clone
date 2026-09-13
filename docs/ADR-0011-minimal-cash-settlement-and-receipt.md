# ADR-0011: Minimal Cash Settlement and Trip Receipt

## Status

Accepted — selected as the next MVP implementation slice.

## Date

2026-09-13

## Context

The Rider ↔ Driver marketplace-to-trip milestone now validates request creation, Driver offers, Rider selection, assignment, Trip start/completion/cancellation, history, and foreground recovery.

That is enough to prove marketplace mechanics, but it is not enough for a credible small pilot. A completed cash ride still needs commercial closure: both sides must be able to see the agreed fare, understand whether cash is still due, and recover a receipt-like record after completion.

The existing marketplace already captures the agreed fare at Rider selection. PR #83 should use that snapshot as the source of truth rather than introducing pricing recalculation, wallets, card payment authorization, cancellation fees, refunds, or no-show policy.

## Decision

Implement a minimal cash-settlement and receipt slice next.

The slice must preserve the current marketplace model:

- The Rider proposes a fare.
- The Driver accepts that fare or submits a counteroffer.
- The Rider explicitly selects one offer.
- The selected offer's fare becomes the agreed fare snapshot for the Trip.
- Trip assignment, start, completion, cancellation, and history remain server-owned.

For the MVP, settlement is limited to cash handling after an assigned Trip reaches completion. The implementation may introduce settlement state, but it must stay small and explicit. The intended business states are:

- `cash_due` — the Trip is complete and the agreed cash fare is still outstanding or unconfirmed.
- `cash_confirmed` — the Driver has confirmed cash collection for the agreed fare.
- `settled` — the Trip has a closed settlement state suitable for Rider/Driver history and receipt display.

PR #83 may collapse `cash_confirmed` and `settled` if the implementation proves that an extra intermediate state adds no MVP value. It must not introduce hidden automatic settlement semantics.

Rider and Driver history should expose receipt-relevant facts from immutable Trip context:

- agreed fare amount and currency;
- service code/name where available;
- pickup and destination coordinates or presentation already accepted by the current product;
- Driver/vehicle context already captured for the assigned Trip;
- Trip status and completion time;
- settlement status.

The Driver completion path may require or immediately follow with a cash-collected confirmation. The exact user interaction belongs to PR #83, but completion and settlement must remain recoverable after app restart or transient network failure.

## Non-goals

This ADR does not authorize:

- Stripe, card processing, wallet balance, payouts, or stored payment methods;
- cancellation fees, refunds, partial payments, tips, promo credits, or no-show policy;
- automated fare calculation, surge pricing, routing-based pricing, tolls, taxes, or commissions;
- background dispatch workers, Redis settlement queues, accounting ledgers, or payout reconciliation;
- broad administrator payment operations;
- Courier/Freight settlement behavior.

Those may become separate decisions only when a concrete vertical slice needs them.

## Consequences

The MVP gains a commercially closed ride loop without committing to a payment platform.

The agreed fare remains the marketplace-selected fare, not a post-trip recalculation.

The settlement model must be small enough for a solo-developer MVP but structured so future online payments can replace or extend the cash settlement mechanism without rewriting Trip history.

PR #83 should be a vertical product slice through persistence, backend APIs, and Flutter UI. It should not start broader payment infrastructure.
