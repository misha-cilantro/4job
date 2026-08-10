# Four Job Fiesta rules, as published

Reference copy of the official rules, so the data files can be checked without
hitting fourjobfiesta.com again.

**Fetched:** 2026-08-09

**Sources**

| Source | URL | Notes |
| --- | --- | --- |
| Registration options | <https://www.fourjobfiesta.com/help.php> | Authoritative and current. Includes every run type. |
| Older help page | <https://www.fourjobfiesta.com/NewHelp.html> | Predates Geyser/Sirocco/Onsen/Haboob. Agrees with help.php where they overlap. |
| Jobs | <https://wiki.fourjobfiesta.com/wiki/index.php?title=Jobs> | Per-crystal job lists. |
| Run types | <https://wiki.fourjobfiesta.com/wiki/index.php?title=Run_types> | Names only; a stub with no pool detail. |
| Duplicate | <https://wiki.fourjobfiesta.com/wiki/index.php?title=Duplicate> | Allow Duplicates scope. |
| Regchaos | <https://wiki.fourjobfiesta.com/wiki/index.php?title=Regchaos> | Legacy run type, retired 2023. |
| Freelancer | <https://wiki.fourjobfiesta.com/wiki/index.php?title=Freelancer> | Nonstandard-job status. |

**Fidelity note.** Passages in quotation marks are verbatim from the source.
Everything else is a condensed restatement, and the wiki pages in particular
were thin. Where the two help pages overlap they agree, which is the basis for
treating the tables below as reliable.

---

## Run types

Each run assigns four jobs, one per slot. A run type decides which crystal
pools each slot may draw from. Quoted text is from help.php.

| Run type | Job 1 | Job 2 | Job 3 | Job 4 |
| --- | --- | --- | --- | --- |
| Regular | Wind | Water | Fire | Earth |
| Typhoon | Wind | Wind, Water | Wind, Water, Fire | Wind, Water, Fire, Earth |
| Geyser | Wind, Water | Wind, Water, Earth | Water, Fire, Earth | **Water, Earth** |
| Sirocco | Wind, Fire | Wind, Water, Fire | Wind, Fire, Earth | Fire, Earth |
| Onsen | Wind, Water | Water, Fire | Water, Fire, Earth | Water, Fire, Earth |
| Haboob | Wind, Earth | Wind, Water, Earth | Wind, Fire, Earth | Fire, Earth |
| Volcano | Wind, Water, Fire, Earth | Water, Fire, Earth | Fire, Earth | Earth |
| Meteor | any | any | any | any |

Verbatim, for the two that matter most:

- Volcano: "Job 1 will come from the Wind, Water, Fire, or Earth Crystal. Job 2
  will come from the Water, Fire, or Earth Crystal. Job 3 will come from the
  Fire or Earth Crystal. Job 4 will come from the Earth Crystal."
- Meteor: "Jobs 1 through 4 will come from _any_ Crystal. Freelancer and Mime
  are also available from any Crystal, and will be available _regardless of Job
  Sets_."
- Geyser: "Job 1 will come from the Wind or Water Crystal. Job 2 will come from
  the Wind, Water, or Earth Crystal. Job 3 will come from the Water, Fire, or
  Earth Crystal. Job 4 will come from the Water or Earth Crystal."

Typhoon leans toward early jobs; Volcano toward later ones. Geyser, Sirocco,
Onsen and Haboob are the four two-element blends.

### A search-result discrepancy, resolved

A web search summary claimed Volcano is "Job 1 from Wind or Water ... Job 4
from Water or Earth" and Meteor is "Job 1 from Wind or Fire". Those are
actually **Geyser** and **Sirocco**. The summariser mislabelled them. Both
help.php and NewHelp.html independently give the Volcano and Meteor rows above,
so the table stands. Recorded here so it doesn't get re-litigated.

### Legacy

- **Regchaos** — introduced 2015, retired 2023. "Replaced by the Volcano run
  type with a Duplicates clause." Legacy page only.
- NewHelp.html lists only Regular, Typhoon, Volcano and Meteor, so the four
  blend types are a later addition.

---

## Job sets

- **Team 750** — jobs that "can break rods or Jobs that are similar in play
  style and theme": White Mage, Black Mage, Blue Mage, Time Mage, Summoner,
  Red Mage, Geomancer, Bard, Chemist, Dancer.
- **Team No 750** — jobs that "cannot break rods and follow within that play
  style and theme": Thief, Monk, Knight, Mystic Knight, Berserker, Ninja,
  Beastmaster, Ranger, Samurai, Dragoon.
- **Team 375** — "This set will assign two Team 750 Jobs and two Team No 750
  Jobs. This set will enable Duplicates."

Ten jobs each; together they cover the 20 crystal jobs exactly, and exclude
Freelancer and Mime.

### Classic and Onion

help.php lists both under **Job Sets**, not as run types.

- **Classic Jobs** — "Jobs 1 through 4 can be any of the following jobs:
  Knight, Thief, Monk, Red Mage, White Mage, Black Mage." All four slots draw
  from the same six-job pool, so Classic replaces the crystal mapping rather
  than filtering it. Nothing in either source says Classic enables or forces
  duplicates.
- **Onion Jobs** — twelve jobs that appeared in Final Fantasy 3, fixed per
  slot: Job 1 Black Mage / Knight / Thief; Job 2 Monk / Red Mage / White Mage;
  Job 3 Bard / Geomancer / Summoner; Job 4 Dragoon / Ninja / Ranger.

---

## Crystal job lists

From the wiki Jobs page. Twenty assignable jobs.

| Crystal | Jobs |
| --- | --- |
| Wind | Knight, Monk, Thief, White Mage, Black Mage, Blue Mage |
| Water | Berserker, Mystic Knight, Time Mage, Summoner, Red Mage |
| Fire | Ninja, Ranger, Beastmaster, Geomancer, Bard |
| Earth | Dragoon, Samurai, Chemist, Dancer |

### Freelancer and Mime are not crystal jobs

- **Freelancer** is listed as "Always Available" — every character starts as
  one, "as the first six jobs are not unlocked until the player reaches the end
  of the Wind Shrine." It "is considered a nonstandard job in that it can only
  be assigned in certain run types, such as Pure Chaos."
- **Mime** is listed under "Special Acquisition", obtained from the Sunken
  Walse Tower.

So neither belongs to a crystal pool. Per the Meteor entry they are "available
from any Crystal, and will be available regardless of Job Sets" — a Job Set
does **not** filter them out.

GBA-only jobs (Gladiator, Cannoneer, Oracle, Necromancer) exist but are outside
the four-slot assignment; Gladiator/Oracle/Cannoneer appear only via the Extra
Jobs modifier.

---

## Modifiers

- **Allow Duplicates** — "If selected, this will allow the same Job to be
  assigned multiple times." Both the wiki Duplicate page and NewHelp.html
  restrict it to **Typhoon, Volcano and Meteor** runs; NewHelp.html says it
  "applies exclusively" to those three. Introduced for the 2023 event. Team 375
  enables it on its own.
- **No Restrictions** — any character may use any assigned job, and you may
  swap which character uses which at any time.
- **Natural Jobs** — Bartz uses Job 1, Lenna Job 2, Faris Job 3, Galuf Job 4;
  Krile uses Job 4 unless Fifth Job is applied.
- **Upgrade Jobs** — all four characters start on Job 1; you may unlock the
  next job at any point, after which earlier jobs are off-limits.
- **Fifth Job** — a fifth job unlocked when Krile joins; retire one earlier job.
- **Extra Jobs** — assigns one Advance job (Gladiator, Oracle or Cannoneer);
  retire one earlier job.
- **Forbidden** — on entering the Void, one job is crossed out and may no
  longer be used.
- **Berserker Risk** — raises the chance of one or more Berserkers, scaling
  with the event's donation total.

---

## How this project's data compares

Checked against `data/runTypes.json`, `data/jobSets.json`, `data/jobPools.json`
and `data/jobs.json` on 2026-08-09.

### Matches

- All eight published run types match slot for slot, Geyser included since the
  fix below.
- Freelancer and Mime sit in their own `special` pool rather than in a crystal
  pool, and bypass job set filtering, as the Meteor rule requires.
- Team 750 and Team No 750 rosters match exactly, all ten jobs each.
- Team 375's two-and-two split and its forcing of Allow Duplicates match.
- The `classic` pool matches the six Classic jobs.
- The four `onion_*` pools match the twelve Onion jobs, slot for slot.
- The base twenty crystal jobs match.

### Fixed after this comparison

1. **Geyser slot 4** was `["fire", "earth"]`; the rule is Water or Earth. It
   also contradicted the project's own description of Geyser as favouring Water
   and Earth. `TestRunTypePoolsMatchPublishedRules` now pins the whole table
   above, so a future data edit that drifts from the rules fails.

2. **Freelancer and Mime were crystal jobs**, sitting in the `wind` and `water`
   pools. They're now a separate `special` pool that every slot may draw from
   when the run type allows them, and they're exempt from job set filtering, so
   Meteor + Team 750 can roll them as the rules require.

3. **Allow Special Jobs was a player toggle on every run type.** It's now fixed
   by the run type: only run types declaring `allowSpecial` can roll Freelancer
   or Mime, matching the rules, where they're assignable only on Meteor (and
   "Pure Chaos" per the Freelancer page). The options step shows the setting for
   information but offers no way to change it.

### Accepted deviations

These differ from the official registration form on purpose. This app builds
bespoke runs, so mirroring the form isn't the goal.

1. **Classic and Onion are run types here, job sets upstream.** Functionally
   close, since both override the crystal mapping anyway, and `noJobSetSelect`
   stops them being combined with a Team set. Worth knowing if a run type and
   job set ever need to compose.

2. **Classic forces Allow Duplicates.** Not stated on either source — Team 375
   is the only selection documented to enable duplicates — but implied, and
   kept deliberately as a house rule. Team 375 and Classic are the two
   selections that force it; `TestDuplicatesForcedWhereRequired` pins both.

3. **Allow Duplicates is offered for every run type.** Upstream it's Typhoon,
   Volcano and Meteor only. Harmless where it's meaningless: run types whose
   slots draw from disjoint pools can't produce a duplicate anyway, so the
   toggle simply has no effect.

4. **Excluded jobs** have no upstream equivalent at all. They're the point of
   the tool — avoiding recent repeats.

5. **Team Vibe Coded** is invented here, not an official job set, and its
   description says so. Every official option filters on one of two axes: which
   crystal, or the 750 split. This one filters on *role*, using the `physical` /
   `magic` / `support` / `combat` tags in `jobs.json`, and pins one slot to each.
   It's the only option that constrains party composition rather than job
   identity. Measured over 20,000 rolls, a plain Normal run leaves you with no
   support job 11.7% of the time and an all-physical party 9.2% of the time;
   with this set both go to zero.

### Modifiers

Implemented:

- **Job restrictions** — No Restrictions, Natural Jobs and Upgrade Jobs, as a
  three-way choice on the options step. The rule is written into each job file:
  Natural names the character who owns that job, Upgrade explains the unlock.
- **Fifth Job** — adds a fifth slot, written to `05_krile.txt` with its
  retire-a-job rules underneath. Under Natural it gives Krile a job of her own
  instead of Galuf's, and the Earth file stops claiming her.
- **Extra Jobs** — adds an Advance job (Gladiator, Oracle or Cannoneer) from a
  new `advance` pool. No crystal and no job set covers these, so nothing but the
  excluded jobs restricts that slot. They unlock **once you have all twelve
  legendary weapons** — that timing came from the project owner, not from any
  page above, which is why it isn't in the quoted sections.
- **Forbidden** — off, rolled now, or left to the player. The rules don't say
  who decides which job is crossed out, so both readings are offered. Rolled
  mode picks one of the run's own assigned jobs and names it in a `_forbidden`
  file; player mode writes the rule and leaves the choice open. Either way it
  takes a job away rather than adding one, so it's not a slot and never changes
  the job count.

Not implemented:

- **Berserker Risk** — weights the roll toward Berserker, "rising with the
  event's donation total". There's no local equivalent of a donation total, so
  there's no faithful version of it. Could be added as a plain weight.

An assumption worth flagging: the rules don't say which crystal a **fifth job**
comes from. It draws from everything the run type can reach across all four of
its slots, and from either side of a counted job set like Team 375, whose four
counted slots are already spent by then.

### Closed: the "Pure Chaos" loose end

The Freelancer page mentions "Pure Chaos" as a run type that can assign
nonstandard jobs, and Regchaos was retired in favour of "Volcano with a
Duplicates clause". Neither "Pure" nor "Chaos" appears on the current help.php
run type list, so there may be run types, or older names for existing ones,
that the pages above don't document.

**Decision: not pursuing this.** Nothing in the project will be built or changed
on the strength of it. The eight run types on the current registration form are
taken as the complete set. Recorded so it doesn't get re-investigated.
