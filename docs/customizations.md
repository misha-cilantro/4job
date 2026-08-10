# What this project does differently

Everything here is this project's own: additions to the official Four Job Fiesta
rules, deliberate departures from them, decisions made where the rules are
silent, and text written for the generated files.

The published rules live in [from-wiki.md](from-wiki.md), which contains nothing
invented here. If you add something to the app that the rules don't specify,
record it in *this* file, not that one.

Last checked against `data/*.json` on 2026-08-10.

---

## Verified against the rules

These match the published rules exactly, and tests keep them that way:

- All eight published run types, slot for slot
  (`TestRunTypePoolsMatchPublishedRules`).
- The twenty crystal jobs (`TestCrystalPoolsMatchPublishedRules`).
- Team 750 and Team No 750, all ten jobs each, together covering the crystal jobs
  exactly once (`TestJobSetPoolsMatchPublishedRules`).
- Team 375's two-and-two split, and its forcing of Allow Duplicates.
- The `classic` pool and the four `onion_*` pools.
- Freelancer and Mime living outside the crystal pools and bypassing job set
  filtering, as the Meteor rule requires.

---

## Fixed after comparing against the rules

Three things the data got wrong, found by reading the sources:

1. **Geyser slot 4** was `["fire", "earth"]`; the rule is Water or Earth. It also
   contradicted the project's own description of Geyser as favouring Water and
   Earth. `TestRunTypePoolsMatchPublishedRules` now pins the whole table, so a
   future data edit that drifts from the rules fails.
2. **Freelancer and Mime were crystal jobs**, sitting in the `wind` and `water`
   pools. They're now a separate `special` pool that every slot may draw from when
   the run type allows them, exempt from job set filtering, so Meteor + Team 750
   can roll them as the rules require.
3. **Allow Special Jobs was a player toggle on every run type.** It's now fixed by
   the run type: only run types declaring `allowSpecial` can roll Freelancer or
   Mime. The options step shows the setting but offers no way to change it.

---

## Deliberate departures

The app builds bespoke runs, so mirroring the official registration form isn't
the goal.

1. **Classic and Onion are run types here, job sets upstream.** Functionally
   close, since both override the crystal mapping anyway, and `noJobSetSelect`
   stops them being combined with a Team set. Worth knowing if a run type and job
   set ever need to compose. It does mean Onion + Team 375 is unavailable despite
   being satisfiable.
2. **Classic forces Allow Duplicates.** Not stated in either source — Team 375 is
   the only selection documented to enable duplicates — but implied, and kept
   deliberately as a house rule. Team 375 and Classic are the two selections that
   force it; `TestDuplicatesForcedWhereRequired` pins both.
3. **Allow Duplicates is offered for every run type.** Upstream it's Typhoon,
   Volcano and Meteor only. Harmless where it's meaningless: run types whose slots
   draw from disjoint pools can't produce a duplicate anyway, so the toggle simply
   has no effect.
4. **Excluded jobs** have no upstream equivalent at all. They're the point of the
   tool — avoiding recent repeats.

---

## Team Vibe Coded

An invented job set, not an official one. Its in-app description says so.

Every official option filters on one of two axes: which crystal, or the 750
split. This one filters on **role**, using the `physical` / `magic` / `support` /
`combat` tags in `jobs.json`, and pins one slot to each. It's the only option that
constrains party *composition* rather than job identity.

Measured over 20,000 rolls:

| | no support | no magic | all physical |
| --- | --- | --- | --- |
| Normal | 11.68% | 5.90% | 9.18% |
| Normal + Team Vibe Coded | 0.00% | 0.00% | 0.79% |
| Normal + Team 375 | 8.35% | 0.00% | 0.00% |

Three things to know about it:

- **The 0.79% all-physical residual is not a bug.** `mystic knight` is tagged both
  physical and magic, and `beastmaster` both physical and support, so the magic
  and support slots can each be filled by a hybrid.
- **On Meteor the guarantee can break.** Special jobs ignore job sets by rule, so
  a Freelancer can land in the slot meant to cover support.
  `TestVibeCodedYieldsToSpecialJobs` documents it.
- **Volcano relaxes about 3.5% of the time.** Its slot 4 is Earth-only and
  `support ∩ earth` is just `chemist`, so if chemist went earlier the roll allows
  a duplicate and says so in a note. Deliberately *not* `forcesAllowDuplicates`:
  the relaxation ladder handles it only when needed and tells you, rather than
  forcing duplicates on all eight run types to fix one.

### Derived tag pools

Team Vibe Coded needs a pool per role. Rather than hand-copy four job lists that
could drift from the tags, `addTagPools` derives one pool per distinct tag at
load, named `tag:<name>`. The prefix avoids colliding with `special` and
`advance`, which are each both a tag *and* a hand-written pool name.
`TestTagPoolsMirrorTheTags` guards the generator.

---

## Modifiers

### Implemented

- **Job restrictions** — No Restrictions, Natural Jobs and Upgrade Jobs, as a
  three-way choice. The rule is written into each job file: Natural names the
  character who owns that job, Upgrade explains the unlock.
- **Fifth Job** — adds a fifth slot, written to `05_krile.txt`. Under Natural it
  gives Krile a job of her own instead of Galuf's, and the Earth file stops
  claiming her.
- **Extra Jobs** — adds an Advance job (Gladiator, Oracle or Cannoneer) from an
  `advance` pool. No crystal and no job set covers these, so nothing but the
  excluded jobs restricts that slot.
- **Forbidden** — off, rolled now, or left to the player.

### Not implemented

- **Berserker Risk** — weights the roll toward Berserker, "rising with the event's
  donation total". There's no local equivalent of a donation total, so there's no
  faithful version of it. Could be added as a plain weight. Dropped by decision,
  not oversight.

---

## Decisions where the rules are silent

Each of these fills a gap listed at the end of [from-wiki.md](from-wiki.md).

- **Which crystal a Fifth Job comes from.** It draws from everything the run type
  can reach across all four of its slots, and from either side of a counted job
  set like Team 375, whose four counted slots are already spent by then.
- **Who decides the Forbidden job.** Both readings are offered. Rolled mode picks
  one of the run's own assigned jobs and names it in a `_forbidden` file; player
  mode writes the rule and leaves the choice open. Either way it takes a job away
  rather than adding one, so it isn't a slot and never changes the job count. A
  rolled Forbidden can land on the fifth or Advance job, since by the Void you
  have all of them.
- **When the Advance jobs unlock.** Once you have all twelve legendary weapons.
  This came from the project owner, not from any page in from-wiki.md.
- **"Pure Chaos."** Not pursuing it. Nothing in the project will be built or
  changed on the strength of it, and the eight run types on the current
  registration form are taken as the complete set. Recorded so it doesn't get
  re-investigated.

---

## Text written for the generated files

None of the wording in a run folder comes from the sources; it's all written
here, in `rules.go`. Two parts are game knowledge rather than fiesta rules, so
they're worth flagging if they ever need correcting:

- **`crystalTriggers`** — where each crystal shatters (Wind Shrine, Walse Tower,
  Karnak Castle, Ronka Ruins), used by `00_instructions.txt` to say when to open
  each file.
- **Unlock timings for the extra jobs** — when Krile joins, and the twelve
  legendary weapons above.

The rule text itself paraphrases the official wording, except the Fifth Job lines,
which are quoted as the project owner supplied them.
