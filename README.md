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
2. **Job set** — Team 750, Team No 750, Team 375, or none. Skipped for run types
   that set their own job pools (Onion, Classic).
3. **Excluded jobs** — anything you'd rather not roll again.
4. **Options** — Allow Duplicates. Forced on by Classic and Team 375. Whether
   Freelancer and Mime can be rolled is fixed by the run type, so it's shown
   here but not adjustable.

Nothing is written to disk until you confirm at the summary. Quitting at any
point writes nothing.

**Keys:** `up`/`down` or `k`/`j` to move, `space` to toggle, `enter` to
continue, `esc` or `backspace` to go back, `r` to start over from the summary,
`q` or `ctrl+c` to quit.

## Output

A folder in the working directory, named from the run's settings plus a
timestamp:

```
Normal-Team 750-excl-3-optX-20260809143000/
  01_wind.txt     white mage
  02_water.txt    red mage
  03_fire.txt     bard
  04_earth.txt    chemist
```

The `-opt` suffix records the modifiers: `D` for duplicates, `S` for special
jobs, `X` for neither.

If your constraints can't all be met — say you excluded every Earth job on a
Normal run — the picker relaxes them one at a time, loosest last, and prints a
note saying what it had to give up rather than failing or hanging.

## Data

The rules live in `data/*.json`, embedded into the binary at build time:

| File | Contents |
| --- | --- |
| `runTypes.json` | Run types and the crystal pools each job slot draws from |
| `jobSets.json` | Team 750 / No 750 / 375 |
| `jobPools.json` | Named job pools: the four crystals, the 750 split, Classic, Onion, and the special jobs |
| `jobs.json` | Every job and its descriptive tags |

Editing them is the intended way to add or adjust a run type. Startup
validation rejects data that can't work: unknown pool or job names, a run type
without exactly four slots, a job set that pins the wrong number of slots, a
run type / job set pairing where some slot has no legal job, and disagreement
between the `special` tag and the special pool.

`from-wiki.md` is a local copy of the official rules with a section comparing
them to these files, including which deviations are deliberate. Check it before
changing the data — and prefer it to re-reading the wiki.

## Layout

| File | Role |
| --- | --- |
| `main.go` | The `run` model and the top-level flow |
| `model.go` | Bubble Tea wizard: steps, keys, views, list scrolling |
| `picker.go` | Rolling jobs, job set restrictions, the relaxation ladder |
| `writer.go` | Creating the run folder |
| `name.go` | Building and sanitising the folder name |
| `data.go` | Loading and validating the embedded JSON |
| `types.go` | Data types and custom JSON unmarshalling |

## Tests

```sh
go test ./...
```

`data_test.go` pins the data files against the rules transcribed in
`from-wiki.md`, so a data edit that drifts from the published tables fails.
`wizard_test.go` drives the Bubble Tea model by feeding it messages directly,
since the TUI needs a real terminal and can't be scripted through stdin.
