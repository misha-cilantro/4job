# Intro

Hamlo! This is me, Misha the human, writing. I wanted a little more flexibility than the
official Four Job Fiesta site provides, and I wanted to learn some go, and I'm
supposed to get LLM-literate for the day job. So, this is the result.

Run it, pick what kind of run you want, exclude jobs you just played with. It'll
make a folder with your run. Open each text file as you reach the appropriate
crystal just like you would on the website.

Please consider still donating to [Four Job Fiesta's charity](https://donate.tiltify.com/236d46a5-eb52-4afd-b11f-516bda3febd1/express/amount) :)

Everything below here is Claude's generated docs.

# 4job

A terminal wizard for building a bespoke [Four Job Fiesta](https://wiki.fourjobfiesta.com/wiki/index.php?title=Main_Page)
run of Final Fantasy V. It rolls four jobs against the run type and job set you
choose, and writes them to a folder of text files — one per crystal, ready to
point an OBS text source at.

The official registration form rolls a run for you and mails it. This does the
same job locally, and adds the things the form doesn't offer: excluding jobs so
you don't repeat a recent run, and mixing options the form keeps apart.

## Usage

```sh
go run .        # or: go build -o 4job.exe . && ./4job.exe
```

The wizard walks four steps, then a summary:

1. **Run type** — Normal, Typhoon, Geyser, Sirocco, Onsen, Haboob, Volcano,
   Meteor, Onion or Classic. Decides which crystal pools each job slot draws
   from.
2. **Job set** — Team 750, Team No 750, Team 375, Team Vibe Coded, or none.
   Skipped for run types that set their own job pools (Onion, Classic).

   *Team Vibe Coded* is a house addition rather than an official set. Where every
   official option filters on which crystal or on the 750 split, this one filters
   on **role**: one slot must be a physical job, one a mage, one a support job,
   and one more a fighter, assigned in a random order. It's the only option that
   constrains party composition — a plain Normal run leaves you healerless about
   one time in nine, and this takes that to zero.
3. **Excluded jobs** — anything you'd rather not roll again.
4. **Options**
   - **Allow Duplicates** — forced on by Classic and Team 375.
   - **Job restrictions** — No Restrictions, Natural Jobs or Upgrade Jobs.
     Cycles with `space`, since the three are mutually exclusive.
   - **Fifth Job** — a fifth job when Krile joins.
   - **Extra Jobs** — one of the GBA Advance jobs.
   - **Forbidden** — a job is crossed out on entering the Void. Cycles between
     off, rolling which job now, and leaving it for you to pick in-game. Takes a
     job away rather than adding one, so it never changes the job count.

   Whether Freelancer and Mime can be rolled is fixed by the run type, so it's
   shown here but not adjustable.

Nothing is written to disk until you confirm at the summary. Quitting at any
point writes nothing.

**Keys:** `up`/`down` or `k`/`j` to move, `space` to toggle, `enter` to
continue, `esc` or `backspace` to go back, `r` to start over from the summary,
`q` or `ctrl+c` to quit.

Colours come from the terminal's own 16-colour palette rather than fixed RGB, so
the wizard fits whatever theme you already run. It drops to plain text on a
terminal without colour, when output is redirected, and when `NO_COLOR` is set.

Text wraps to the terminal width, and the two long lists scroll to its height —
both recomputed as you resize. Very narrow or very short terminals stop shrinking
at a floor (`minTextWidth`, `minListRows`) and overflow instead, on the grounds
that a column of single syllables is worse than a little scrollback.

## Output

A folder in the working directory, named from the run's settings plus a
timestamp. Each job goes in its own file, with the job name on the first line so
a one-line stream source still reads correctly:

```
Normal-Team 750-excl-1-optN5AF-20260809143000/
  00_instructions.txt  which file to open when
  00_rules.txt      the run's settings and its whole-run rules
  01_wind.txt       blue mage
                    Bartz must always be a blue mage.
  02_water.txt      summoner
                    Lenna must always be a summoner.
  03_fire.txt       geomancer
  04_earth.txt      chemist
  05_krile.txt      red mage           (Fifth Job)
  06_advance.txt    gladiator          (Extra Jobs)
  07_forbidden.txt  gladiator          (Forbidden, rolled)
                    Crossed out on entering the Void. You may no longer use it.
```

`00_instructions.txt` lists exactly the files this run has and when to open each
one — at the Wind Shrine, at Walse Tower, when Krile joins, on entering the Void
— so you can roll a run now and play it blind later.

Nothing spoils the roll. The app never prints your jobs to the terminal, and
`00_rules.txt` describes the run without naming a single assigned job, so the only
way to see one is to open its file. Open them early if you'd rather know.

Rules are written under a job only when one actually applies to it, so a plain
run with no advanced options writes just the job name, exactly as before.
Numbering is sequential, so an Advance job is `05_advance` on its own and
`06_advance` alongside a fifth job, and the forbidden file follows the last job.

With Forbidden set to roll now, the forbidden file names the job and the job's
own file gains a line saying it gets crossed out. Set to *by player*, the file
reads `set by player` with the rule underneath, and nothing is decided for you.

The `-opt` suffix records the modifiers: `D` duplicates, `S` special jobs,
`N` natural, `U` upgrade, `5` fifth job, `A` Advance job, `FR` forbidden rolled,
`F` forbidden by player, or `X` for a run with none of them.

If your constraints can't all be met — say you excluded every Earth job on a
Normal run — the picker relaxes them one at a time, loosest last, and prints a
note saying what it had to give up rather than failing or hanging.

## Data

The rules live in `data/*.json`, embedded into the binary at build time:

| File | Contents |
| --- | --- |
| `runTypes.json` | Run types and the crystal pools each job slot draws from |
| `jobSets.json` | Team 750 / No 750 / 375 / Vibe Coded |
| `jobPools.json` | Named job pools: the four crystals, the 750 split, Classic, Onion, the special jobs and the Advance jobs |
| `jobs.json` | Every job and its descriptive tags |

A pool per tag (`tag:physical`, `tag:magic`, …) is derived from `jobs.json` at
load rather than written by hand, so the tags can't drift from the pools built on
them. That's what Team Vibe Coded draws on.

Editing these files is the intended way to add or adjust a run type. Startup
validation rejects data that can't work: unknown pool or job names, a run type
without exactly four slots, a job set that pins the wrong number of slots, a
run type / job set pairing where some slot has no legal job, and disagreement
between the `special`/`advance` tags and their pools.

## Docs

Two files under `docs/`, split by where the information came from. The split is
worth preserving: it's what makes the first file usable as an authority.

| File | Contents |
| --- | --- |
| [`docs/from-wiki.md`](docs/from-wiki.md) | The official rules, copied locally with sources and fetch date. Nothing invented here. Ends with the gaps the sources don't settle. |
| [`docs/customizations.md`](docs/customizations.md) | Everything this project adds or changes: house rules, deliberate departures, decisions made where the rules are silent, and the text written into the generated files. |

Check both before changing `data/`, and prefer them to re-reading the wiki. If you
add something the rules don't specify, record it in `customizations.md`.

## Layout

| File | Role |
| --- | --- |
| `main.go` | The `run` model and the top-level flow |
| `model.go` | Bubble Tea wizard: steps, keys, views, list scrolling |
| `picker.go` | Rolling jobs, job set restrictions, the relaxation ladder |
| `rules.go` | Job restriction modes and the rule text written into the files |
| `styles.go` | The handful of lipgloss styles the views share |
| `writer.go` | Creating the run folder and its files |
| `name.go` | Building and sanitising the folder name |
| `data.go` | Loading and validating the embedded JSON |
| `types.go` | Data types and custom JSON unmarshalling |

## Tests

```sh
go test ./...
```

`data_test.go` pins the data files against the rules transcribed in
[`docs/from-wiki.md`](docs/from-wiki.md), so a data edit that drifts from the
published tables fails. `wizard_test.go` drives the Bubble Tea model by feeding it
messages directly, since the TUI needs a real terminal and can't be scripted
through stdin.
