# When to Decompose Markup into Components

This framework composes templates freely, so mechanics are never the constraint.
This guide is the judgment call: when a chunk of markup should become its own
component, and when it should stay inline.

## Principle

Aim for high cohesion (everything in a component serves one purpose) and low
coupling (a component depends little on others). The operational test:

> A component should have **one reason to change.**

Two parts that change for different reasons are a seam. A chunk that only ever
changes for one reason is already cohesive. Decompose to resolve a real problem,
not to hit a line count — size is a symptom to investigate, not a cause.

## Extract when

- **It's reused, or will be.** The strongest trigger. One source of truth beats
  duplicated markup that drifts out of sync.
- **It has one distinct, nameable responsibility.** A specific "and"-free name
  (`PriceBadge`, `NavBreadcrumbs`) is evidence of a cohesive unit.
- **It owns its own behavior** — local state, interactivity, or logic that would
  otherwise tangle through the parent.
- **Extraction makes the parent read like an outline.** Readability alone counts.
- **It's a non-trivial list item.** Per-item markup in a loop is reused by
  construction; give it one responsibility.

Size (past one screen, or ~150–200 lines) is a cue to *look for* one of the
above — split along the seam you find, not at the line count.

## Keep inline when

- The chunk is small, static, and used once — extraction adds only indirection.
- The only motive is length, with no seam — splitting spreads one responsibility
  across two files, worsening cohesion.
- Extracting would thread many variables in and events back out — that coupling
  survives the move and just turns verbose. Redraw the boundary or leave it.
- The only available name is vague (`Section2`, `ContentWrapper`) — no real
  responsibility exists to carve.
- Reuse is speculative — wait for the second use, then extract.

## Boundary check

Once you've decided to extract, validate the cut by its interface — the coupling
that crosses the new boundary. Moving a chunk into its own component increases
distance, so the coupling across the seam must be weak for the move to pay off:

- **Small and explicit** (a few well-named inputs) → good cut, proceed.
- **Large, or dependent on ordering or shared implicit state** → the boundary is
  wrong. Redraw it, or keep the markup inline.

Strong coupling stretched across a distance is worse than strong coupling kept
local. When two sides stay intertwined, keep them together.

## Procedure

1. Reused or will be? → extract.
2. One distinct, nameable responsibility or its own behavior? → candidate;
   run the boundary check.
3. Just big? → find a seam and split along it; no seam → inline.
4. Boundary check passes (small, explicit interface)? → extract; else inline.
5. In doubt → inline. A small cohesive chunk is cheap to split later; an
   over-fragmented tree is expensive to untangle.