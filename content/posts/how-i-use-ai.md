---
title: "How I Use AI"
date: 2026-08-17
description: "How I configure agent-skills: keeping things harness-agnostic, strict operational rules, MCP servers, and repeatable skill pipelines."
tags: [ai, tooling]
---

Most chat around using AI for engineering seems to land on two extremes:
either people claiming agents will replace developers by next month, or
complete cynicism because a model botched a regex. Neither feels particularly
helpful to be honest.

I use coding agents every single day across multiple codebases. I don't treat
them like magical answers to everything, and I definitely don't let them dump
unvetted code into my repositories. For me, an agent is essentially a very fast,
enthusiastic junior engineer who happens to be a bit careless if you leave them
unsupervised. That means giving them strict boundaries, making TDD
non-negotiable, and having zero patience for apologies or boilerplate slop.

All of this lives in my
[`agent-skills`](https://github.com/Integralist/agent-skills) repo. Here is a
look at the mental model, the setup, and how work actually gets done.

## 🔌 Don't Marry Your Harness (Stay Agnostic)

The CLI agent ecosystem moves ridiculous fast. Every few months there's another
tool or harness turning up: Claude Code, Pi, Gemini CLI, OpenCode, Copilot CLI,
and whatever else launched this morning. Tying your prompts, custom tools, and
habits to one specific harness is a bad idea. The second you want to switch (or
the project dies), you lose everything.

I say this from painful personal experience. I was once so invested in OpenCode
that I [forked it to fix a bunch of
issues](https://github.com/Integralist/opencode/blob/custom-features/FORK.md#changelog),
adding prompt history search, subagent cost tracking, and skill autocomplete. I
tried contributing those changes back upstream. But with hundreds of PRs
sitting stale and an unresponsive maintainer, none of it got merged. It wasn't
for lack of trying, but I quickly realised OpenCode had too many architectural
headaches under the hood and wasn't going to work out long term.

So I moved over to the Pi harness instead. Naturally, I still wanted deeper
capabilities like subagent orchestration, better terminal telemetry, transcript
highlighting, side conversations, inline skill argument hints, declarative
event hooks, and in-place context shaking. So I built extensions for them:
[`pi-statusbar`](https://github.com/Integralist/pi-statusbar),
[`pi-subagents`](https://github.com/Integralist/pi-subagents),
[`pi-btw`](https://github.com/Integralist/pi-btw),
[`pi-transcript-enhancer`](https://github.com/Integralist/pi-transcript-enhancer),
[`pi-skill-argument-hints`](https://github.com/Integralist/pi-skill-argument-hints),
[`pi-event-hooks`](https://github.com/Integralist/pi-event-hooks),
and
[`pi-context-shake`](https://github.com/Integralist/pi-context-shake).

There is an important distinction to make here, though. Right now, the industry
doesn't have a cross-harness standard for runtime extensions. If you want rich
runtime features like subagents or custom terminal UI, you inevitably have to
write against whatever plugin API your harness provides. If I switch harnesses
down the road, that glue code is what I'd have to adapt.

What *can* be standardised, however, is the actual intelligence: the prompt
logic, workflows, and operational conventions. The emerging "Skills" format
gives us a portable way to do that. By keeping the core skills and conventions
strictly harness-agnostic, the way I actually work isn't locked into Pi, Claude
Code, or anything else.

My `agent-skills` repo reflects this separation with a single source of truth:

```txt
.agents/                            # Canonical skills + conventions
├── AGENTS.md                       # Shared operational conventions
└── skills/                         # Canonical skills (1 dir per skill)
    └── shared/                     # Shared formats and policies

.claude/                            # Claude Code adapter
├── CLAUDE.md                       # Pointer to ~/.agents/AGENTS.md
├── rules/                          # Auto-generated path-scoped rules
└── skills -> ../.agents/skills     # Symlink to canonical skills

.pi/agent/                          # Pi harness configuration
├── settings.json                   # Defaults, models, and packages
└── themes/nord-contrast.json       # Custom theme

mcp/google-workspace/               # Bundled Google Workspace MCP server
```

The core directory is `.agents/skills/`. Every skill is written using generic
instructions ("prompt the user", "spawn a subagent", "read the file"). Claude
Code gets access via a symlink (`.claude/skills -> ../.agents/skills`), while Pi
or Gemini CLI load them straight from `~/.agents/skills/`.

To stop skills turning into bloated, harness-specific walls of text, I enforce
strict authoring rules via `writing-for-agents` and use `skill-rewrite` to keep
each skill short enough to read in one pass. Skill descriptions stick strictly
to trigger conditions (when to load the skill, not a duplicate summary of how it
works), avoid harness-only shell-injection syntax, and extract shared contracts
(`SPEC-FORMAT.md`, `ADR-FORMAT.md`, `PR-DESCRIPTION.md`, `BUILD-ITEM.md`) into
`.agents/skills/shared/` so linked workflows never drift.

When Claude Code supports handy features like path-scoped rules in
`.claude/rules/`, I generate them automatically from the canonical conventions
skills (`conventions-go`, `conventions-python`, and so on) via `make rules`.

Similarly, Claude Code supports an `argument-hint` frontmatter key in
`SKILL.md` to show syntax hints when typing slash commands (such as
`[--edit] [--days N]` for `talking-points`). Pi does not display these
natively, so I wrote
[`pi-skill-argument-hints`](https://github.com/Integralist/pi-skill-argument-hints)
to bridge the gap. When typing `/skill:<name>`, it reads the frontmatter
verbatim and renders a muted inline hint right after the cursor that auto-hides
the moment you type an argument or space.

The same story played out with lifecycle hooks. Claude Code supports
declarative shell hooks (`PreToolUse`, `PostToolUse`, `Notification`) in JSON,
making it trivial to block dangerous commands, format files after edits, or pop
desktop alerts. Pi has a powerful event API (`pi.on(...)`), but writing a full
TypeScript extension just to run a one-line bash script is overkill. I built
[`pi-event-hooks`](https://github.com/Integralist/pi-event-hooks) to bring
declarative JSON hooks to Pi. It maps both native Pi events and Claude Code
aliases onto shell commands from `~/.pi/agent/hooks.json` or `.pi/hooks.json`.
Hooks can block risky calls with exit code `2` (like `git push --force` or
touching `.env`), mutate arguments on the fly, or trigger formatters on tool
completion without maintaining separate extensions.

```bash
# Install everything across all harnesses cleanly
make install
```

`make install` uses
[`scripts/op-inject.sh`](https://github.com/Integralist/agent-skills/blob/main/scripts/op-inject.sh)
to inject 1Password secrets for API keys and MCP configs at install time. If
1Password isn't signed in, it quietly skips those injections instead of
blowing up.

## 📜 Ground Rules (Stop the Fluff)

Left to their own devices, agents default to people-pleasing nonsense. They
apologise constantly, hallucinate libraries, churn out paragraphs of
throat-clearing filler, and edit files nobody asked them to touch. My global
`AGENTS.md` sets the boundaries so we don't have to have that argument every
session.

### 💬 1. Communication and Tone

- **Cut the sycophancy:** Drop the "I would be delighted to help with that!"
  nonsense. Lead with the direct answer.
- **Keep it brief:** Shortest complete response wins. Cap lists at around five
  items and group them by priority.
- **Bound the steps:** If a task takes multiple steps, number them upfront so I
  know where we are. My
  [`pi-transcript-enhancer`](https://github.com/Integralist/pi-transcript-enhancer)
  extension parses progress labels (like `Step X/Y: <summary>`) and renders them
  as high-contrast gold blocks in the TUI, keeping long sessions scannable at a
  glance without modifying the raw message sent to the model.
- **No robotic fluff:** Drop literary flourishes and hedging throat-clearing.
  Keep the tone warm, plainspoken, and peer-to-peer.
- **Load-bearing emoji only:** Emoji are fine if they signal status faster than
  words (✅ pass, ❌ fail, ⚠️ caution). Otherwise keep them out of running
  prose and code.

### 🧪 2. Engineering Discipline & TDD

- **Strict TDD:** No production code without a failing test first. Stub first,
  prove failure on assertion (not compilation failure), write the minimum code
  to pass, and delete assertions that survive an inverted requirement.
- **Simplicity over layers:** I prefer solving problems by removing components
  rather than stacking abstractions or wrapper functions.
- **Cite actual sources:** Never guess an API signature or config header from
  vague memory. Cite the file and line number
  (`internal/parser/frontmatter.go:42`), or explicitly flag it as an unverified
  assumption.

### ✋ 3. Sceptical Code Edits

- **Ask before editing:** An inquiry is just an inquiry, not an open invite to
  start hacking on files. The agent has to propose the diff in chat and wait for
  approval before running edit tools.
- **Summarise large diffs:** Anything over 40 lines gets a one-line summary
  first, rather than dumping a massive wall of code into chat unprompted.

## ⚙️ The Skill Pipeline: How Work Actually Gets Done

A common mistake is asking an agent to "build feature X" in a single prompt.
That almost always produces over-engineered junk (and a massive headache to
debug). Instead, I break work down into a pipeline of smaller, specialised
skills.

```txt
Product track:     pdd (project → discovery → design) ─┐
                                                       ├─→ architect (spec → grill → plan) → to-tasks → next-task / next-slice
Engineering track: architect (bootstrap → research) ───┘
```

### 🕵️ Phase 1: Interrogation (Poking Holes First)

Before touching code, the idea needs stress-testing.

- **`clarify`**: Digs into what you actually want when a prompt is too vague. It
  asks targeted questions to remove ambiguity upfront.
- **`grilling`** (and `grill-me`): Puts the idea through an adversarial
  interrogation. The agent relentlessly challenges your assumptions, edge
  cases, and architecture choices across the design tree.
- **`perspectives` & `decide`**: Evaluates alternatives through structured
  lenses (risks, benefits, costs) and produces a durable decision memo or ADR
  (Architecture Decision Record).
- **`consensus`**: Drives cross-model debate through gated discussion rounds,
  preserving healthy dissent when evaluating high-stakes architectural choices.
- **`arena`**: Runs parallel candidate attempts across multiple models for
  tricky pieces where a single shot risks locking in the wrong shape. It scores
  candidates against a concrete rubric via an independent cross-judge, picks a
  base, and borrows the best bits from the runners-up.

### 📐 Phase 2: Architecture and Planning (Before the Code Rot)

Once an idea survives the initial interrogation, it needs structuring. But not
all work starts from the same place. Some initiatives are cross-functional,
needing agreement between Product and Engineering on scope, milestones, and
system design. Others are purely engineering-led problems where we already know
we need to build something, but need to research the shape and plan the work.

My setup splits this into two distinct tracks: `pdd` for cross-functional
initiatives, and `architect` for engineering-led features.

#### Cross-functional alignment: `pdd`

When Product and Engineering need to agree before anyone touches code, I use
`pdd` (Project, Discovery, Design). It treats governance as a series of three
explicit, human-in-the-loop approval gates:

1. **Project (`project.md`)**: Agrees on what the initiative actually is, why
   it matters, and when milestones should land.
1. **Discovery (`discovery.md`)**: Evaluates solution directions and trade-offs
   for a milestone without committing to implementation details.
1. **Design (`design.md`)**: Documents the approved system-level architecture
   thoroughly enough for Product, Engineering, and reviewers to sign off.

Unlike most agent skills that run through to the end autonomously, `pdd` stops
dead at each gate (`Awaiting approval`). It refuses to advance to Discovery
until Project is approved, and will not touch Design until Discovery is signed
off.

Crucially, `design.md` is not an implementation plan (`plan.md`). It answers
what system solution is being approved, not how an engineer will slice the pull
requests. Once Design is approved, `pdd` hands off to `architect` starting at
its Spec step, skipping bootstrap and research so both tracks share one
engineering chain without duplicating the discovery you just agreed upon.

#### Engineering-led coordination: `architect`

When an initiative is driven entirely by engineering, running through a
three-stage product governance dance is overkill. That is where `architect`
comes in.

Where `pdd` is a gated approval workflow for cross-functional consensus,
`architect` is a thin six-step coordinator for engineering delivery. It
sequences other skills in a single `projects/<yyyy-mm-dd>-<slug>/` folder:
bootstrapping project rules (`agents-md`, plus `clarify` if the idea is vague),
deep research (`research`), writing the functional specification (`to-spec`),
adversarial stress-testing (`grill-with-docs`), implementation planning
(`to-plan` followed by `to-adr` to record genuine architectural decisions, and
an optional `to-prd` when the project didn't start from a PDD Design), and
finally handing off to execution.

If `pdd` is about deciding what to build with Product, `architect` is about
taking an engineering idea (or an approved PDD Design) and generating the
technical artefacts needed to build it safely.

#### The engineering handoff: specs, plans, and tasks

Regardless of whether an initiative starts in `pdd` or `architect`, both tracks
converge on the same engineering delivery pipeline under
`projects/<yyyy-mm-dd>-<slug>/`:

- **`to-spec` & `spec-delta`**: `to-spec` writes `spec.md` with user stories
  (framed with `voice-product`), Gherkin acceptance criteria, and testing seams.
  Current truth lives in `docs/specs/<capability>.md` and is validated in CI;
  whenever a change alters existing behaviour, `to-spec` and `spec-delta` record
  a `Behavioural Delta` against the pinned baseline so the project folder keeps
  the history while the living spec moves on.
- **`to-plan`**: Breaks the specification (or a session's discussion when a spec
  isn't needed) into vertical implementation slices (`plan.md`) with `Blocked by`
  dependencies, exact `Consumes`/`Produces` interface signatures, verifiable
  outcome milestones (`M1.1`, `M1.2` rather than step-by-step tasks), and PR
  delivery groupings.
- **`to-tasks`**: Compiles a single plan slice just-in-time against live code
  into a mechanical, TDD-shaped task list (`tasks.md` or `tasks-slice-N.md`)
  with verbatim stubs, failing tests (`red`), implementations (`green`),
  runnable verification commands, `Supports:` links back to plan milestones, and
  a closing integration task that runs the slice's Gherkin scenarios without
  mocks.

#### Automating the handoff: `execute-plan`

Until now, I've described a manual loop: run `to-tasks` for a slice, then start
`next-task` or `next-slice`, and repeat. It works, but I'm the one joining the
steps up and keeping track of what comes next.

[`execute-plan`](https://github.com/Integralist/agent-skills/blob/4601fe188be2b7a4001c40197eb533491f01570d/.agents/skills/execute-plan/SKILL.md)
turns that choreography into one coordinated workflow. I give it an existing
plan and keep my main conversation as the coordinator so its context window
lasts the whole plan. For each unblocked slice, a premium subagent follows
`to-tasks` to write the task list against live code, and a cheap subagent
implements one task at a time (both staying on the session's provider and
changing only the model and effort). The coordinator inspects each diff, runs
the checks, diagnoses failures, and commits verified tasks before moving on. It
accepts each slice against its milestone outcomes (running a read-only review on
risky slices), finishes with a requirement-by-requirement audit against the plan
and spec, and logs task-to-commit mappings in `execution.md` so resuming never
depends on chat history.

The value isn't handing decisions to a cheaper model. It's avoiding the manual
skill-to-skill handoffs while keeping verification and acceptance with the
coordinator. I do give up some of the minute-by-minute control, and when I'm
working manually I occasionally course-correct code to match my conventions or
preferences. In practice, those tend to be small syntax or style adjustments,
not problems with the code or the planned approach. So I can let the task loop
run, review the final code and complete commit list, suggest any small edits,
then decide whether to push or open a PR. That makes the overall flow faster
for me.

By default it confirms the setup and asks when design questions come up. Passing
`auto` (`/skill:execute-plan <plan-path> auto`) lets it run unattended through
all slices and the final audit, logging each non-trivial decision in
`execution.md`. Both modes still stop hard if staging is unsafe, three fixes
fail on the same problem, the worker model can't be verified, the provider runs
out of credits, a commit would land on `main`/`master` without approval, or the
PDD Design isn't approved. It never pushes or opens PRs, and I can still run the
individual skills manually whenever I want to stay closer to each step.

### 🔨 Phase 3: Building (Small, Mechanical Slices)

With tasks compiled, execution becomes fast and predictable.

- **`next-task` & `next-slice`**: Both work in the main thread on a `moderate`
  model tier using a shared test-first procedure (`BUILD-ITEM.md`), whether
  running from a compiled task list or directly from plan milestones.
  `next-task` builds and verifies the next ready item, while `next-slice` works
  through an entire slice, verifies every milestone outcome against the full
  test suite, and offers one commit at the end.
- **`conventions-*` & `go-testing`**: Enforces strict language conventions
  (`conventions-go`, `conventions-python`, `conventions-sql`, and others) and
  testing discipline (table-driven tests, `t.Context()`, bounded waits, public
  API testing, `godog` Gherkin acceptance test wiring, and zero unkeyed struct
  literals).
- **`unslop`**: Strips AI tells, buzzwords, puffery, and robotic cadence from
  documentation and prose to restore a clean, authentic human voice.
- **`cleanup`**: Runs a background audit hunting for AI-generated clutter, dead
  code, and unnecessary abstractions.
- **`precedent`**: Audits newly written code against existing conventions in the
  codebase, flagging any inconsistencies in naming, error handling, or API
  signatures.

### 🧐 Phase 4: Sceptical Review (Don't Be a Pushover)

Reviewing code with AI works both ways: having agents review code, and
critically evaluating the feedback agents give you.

- **`code-review`**: Gathers the diff and intent documents (`Execution-Plan`
  commit trailers, active project specs and plans, and living specs in
  `docs/specs/`), then runs parallel subagents across up to five profiles
  (behaviour and tests, security, reliability and rollout, maintainability and
  conventions, plus spec and plan adherence). Parallel read-only verifier
  subagents then challenge every finding and drop anything they can't prove.
  Every reviewer and verifier runs on the session model or better, because a
  missed defect costs more than the tokens.
- **`is-it-safe`**: Rather than general reviews or speculative risk matrices,
  this runs an adversarial impact analysis answering one question: what is the
  worst credible outcome if this exact diff is merged? It uses an independent
  max-effort subagent pass to trace a concrete failure path (trigger, changed
  behaviour, blast radius) and test safeguards before returning a clear verdict
  (🟢 `PROCEED`, 🟡 `HOLD`, 🔴 `STOP`). It catches nasty blast-radius surprises
  so changes ship safely without relying solely on green tests.
- **`code-review-feedback` & `security-review-feedback`**: When an AI reviewer
  (or static analyser) flags an issue, you shouldn't reflexively accept it.
  These skills force the agent to evaluate the claim with technical rigour. Is
  the vulnerability actually reachable? Is the suggested refactor introducing
  hidden complexity? If a suggestion is rubbish, reject it with proof.
- **`explain-code`**:
  Not every question about a PR or diff is a request for a review. Sometimes I
  just want to understand what changed, why it matters, and how it fits into an
  unfamiliar system. The skill gives newcomers a concise Problem/Solution
  walkthrough of a PR, branch, commit, or scoped code, with a small Mermaid
  diagram when there’s a useful flow to show and references to the sources. It
  explains behaviour and purpose rather than hunting for defects or judging
  whether a change should merge. For a PR, it can also draft a clarification
  comment for the author, but posting needs separate approval.

### 🚢 Phase 5: Shipping Without the Drama

When the code is tested and clean, shipping is a deterministic, mechanical step.

- **`branch`**: Cuts a feature branch using session context and standard naming.
- **`commit`**: Stages and groups files sensibly, enforcing titles that explain
  why the change matters rather than listing raw mechanical steps.
- **`open-pr`**: Generates a ready-for-review pull request with clear Problem
  and Solution sections, consequence-driven titles, and load-bearing emoji from
  a shared `PR-DESCRIPTION.md` contract.
- **`stacked-prs`**: Coordinates dependent PR chains with the official
  `gh stack` CLI extension, opening each layer ready for review (`--auto --open`)
  with the same shared PR description rules.
- **`bcp`**: Thin coordinator that runs `branch`, `commit`, and `open-pr` (or
  hands off to `stacked-prs` when `gh stack view` finds an active stack) in one
  command.

### 🧭 Choosing the Right Analysis Skill

Different engineering problems call for different analytical lenses:

| Skill              | Use when                                                                                   | Primary output                                               |
| ------------------ | ------------------------------------------------------------------------------------------ | ------------------------------------------------------------ |
| **`perspectives`** | Brainstorming or running a "what are we missing?" pass                                     | Multi-perspective analysis                                   |
| **`decide`**       | Choosing between consequential engineering options                                         | Durable decision memo / ADR                                  |
| **`consensus`**    | A complex design or implementation needs independent cross-model review and approval gates | Reviewed assessment or implementation with dissent preserved |
| **`arena`**        | Non-trivial artefact where one attempt risks the wrong shape                               | Synthesised multi-model artefact                             |
| **`precedent`**    | Ensuring new code matches existing codebase conventions                                    | Divergence report citing peer patterns                       |
| **`code-review`**  | Code or diff exists and defects must be identified                                         | Verified findings across subagents                           |
| **`is-it-safe`**   | Assessing the worst credible failure mode and blast radius before merge                    | Adversarial impact report with a PROCEED/HOLD/STOP verdict   |

## 🔍 Tight Feedback Loops with Crit and Hunk

One of the biggest friction points with AI workflows is the review loop.
Normally, you're either squinting at terminal diffs trying to explain in chat
which line needs changing, or you're pushing half-baked branches up to GitHub
just to use their pull request UI. Both options are clunky. Terminal diffs are
annoying to comment on with precision, and pushing to GitHub drags remote
latency and premature commits into what should be a fast local experiment.

This is where [Crit](https://github.com/tomasz-tomczyk/crit) fits in. Crit is a
lightweight, local review tool that runs in your browser. When an agent finishes
drafting a plan, generating a spec, or writing a chunk of code, I run `crit` (or
the agent triggers it via my `crit` skill). It opens an interactive review UI in
the browser where I can see the full diff, highlight specific lines, leave
inline comments, file-level notes, or general feedback, just like a GitHub PR
review, but completely local.

```bash
crit                           # branch diff against main
crit <plan-file>               # review an implementation plan or spec
crit --pr <num|url>            # fetch and review a GitHub PR locally
```

The real beauty is how it closes the loop with the agent. Crit stores review
feedback in a structured local JSON file. The agent reads that file, finds any
unresolved comments, fixes the code or plan, and replies directly to each
comment with what it changed. Once the edits are done, the agent re-runs `crit`
to trigger the next review round.

We can iterate through three or four review rounds locally in minutes, without
pushing a single commit upstream. And when everything finally looks right, `crit push` can sync those local comments straight up to the GitHub PR review if
needed. It keeps the human firmly in the driving seat without breaking the flow.

That said, I actually prefer a terminal experience over a browser UI, so
nowadays I tend to use [Hunk](https://www.hunk.dev/) instead. It gives me that
same local review loop without ever leaving my terminal window.

To wire it into my agent workflow, I built a
[`hunk`](https://github.com/Integralist/agent-skills/blob/cc1bf18b5a6acb1b448f550f48a02c36cdfd51f9/.agents/skills/hunk/SKILL.md)
skill that drives a live
`hunk diff` TUI session through `hunk session` CLI commands. I keep ownership of
the interactive terminal window while the agent reloads it onto whatever I'm
reviewing: uncommitted or staged changes, a branch diff, a specific commit, or a
remote GitHub PR (for cross-repo PRs, it drops the unified diff into
`.git/hunk-review-$SESSION_ID.patch` so concurrent sessions don't clobber each
other and the working tree stays clean).

When the agent reviews a changeset, it batch-applies inline notes, scrolls the
TUI to the first stop, and highlights the exact character span I should look at.
I can step through findings by replying `next` or `prev` in chat, or press `c`
in Hunk to leave my own inline notes. Running `/hunk comments` tells the agent
to read my threads, make any approved code edits, reply directly under each
note, and refresh the view in place. When I'm done, `/hunk done` clears the
session notes, prunes any orphaned patch files, and restores the working-tree
diff.

## 🧑‍🏫 Teaching with Slides (Beyond Code)

Engineering isn't just churning out code. A good chunk of the job is explaining
complex systems to peers, onboarding team members, or helping someone grasp an
unfamiliar concept without overwhelming them.

That is where [`teach-with-slides`](https://github.com/Integralist/agent-skills/blob/main/.agents/skills/teach-with-slides/SKILL.md)
comes in. Instead of spitting out a dry wall of text or generic bullet points,
it builds a structured, beginner-friendly learning deck. It establishes a
proper learning arc, starting with a relatable hook, introducing the core
mental model, and breaking the topic down into digestible steps.

It also handles the presentation format depending on what I need: HTML for a
quick browser deck, PPTX or Google Slides when it needs editing, or PDF for
sharing. It keeps slides focused on a single idea, leans heavily on Mermaid
diagrams to show relationships rather than describing them, and applies clean
visual themes (a warm Claude-inspired palette or an informal Fastly-inspired
style for colleagues).

## 💬 Prepping Retros and 1:1s (Recency Bias Sucks)

Another awkward part of engineering is preparing for team retrospectives and
one-to-ones. Most of us rely on whatever annoyed us over the last forty-eight
hours, meaning chronic friction gets forgotten while yesterday's hiccup takes
centre stage. The real trail is in Slack, scattered across team channels and
DMs.

That is where [`talking-points`](https://github.com/Integralist/agent-skills/blob/328bc82b4874e50c2c1bbc3b8b25ee50e595fa3d/.agents/skills/talking-points/SKILL.md)
comes in. It uses a Slack MCP connector to search messages, team channels, and
DMs across a configurable lookback window (defaulting to two weeks). Rather
than dumping raw hits, it clusters messages by topic and scores them by
frequency, breadth across teammates, recency, and emotional intensity
(weighting your own complaints highest as your strongest signal).

Crucially, it tags the audience for each item (`retro` for public channel
topics, `1:1` for private DMs or personal behaviour). It prints a ranked
terminal summary with the signal and a concrete proposal to bring, while keeping
everything local: basic config on disk, messages in chat, DMs paraphrased, and
credentials redacted on the fly.

## 🌐 MCP & Wiring Up Context

Agents are only as useful as the context you feed them. I use a handful of MCP
servers to bridge the gap between local code and the tools I use every day:

- **Google Workspace MCP**: Bundled locally in `mcp/google-workspace/` (an
  unmodified build of upstream `gemini-cli-extensions/workspace`). Gives the
  agent secure access to Calendar, Drive, Docs, Sheets, and Gmail via local
  OAuth.
- **Atlassian MCP**: Connects to Jira and Confluence using the modern MCP
  Streamable HTTP transport via `mcp-remote` with `--transport http-only`.
- **Language Servers (`gopls`)**: Supplies real-time compiler diagnostics and
  symbol definitions directly to the agent.
- **Context7**: Live documentation indexing and retrieval for up-to-date
  third-party libraries.
- **Slack MCP**: Searches public channels, team channels, and DMs to retrieve
  conversations, decisions, and recurring friction for `talking-points`.

> [!NOTE]
> All sensitive endpoints and API keys are injected dynamically via 1Password
> CLI templates (`.claude.json.tmpl`, `.copilot/mcp-config.json.tmpl`), ensuring
> no secrets ever leak into version control.

## 💰 Keeping Costs Sane

Running every simple command through top-tier frontier models is slow and
expensive. My setup enforces a bit of financial and token discipline:

- **Model Tiering**: Subagent models are chosen by what a miss would cost.
  Mechanical work (searching files, bulk reads, formatting, drafting prose)
  runs on the cheapest tier, building from a finished task list (`next-task`,
  `next-slice`) uses a moderate tier, and judgment-heavy work (code and security
  review, root-cause diagnosis, architecture, and grilling) stays on the session
  model or better.
- **[`pi-subagents`](https://github.com/Integralist/pi-subagents)**: An
  extension I built to run isolated subagents in the same process. Each subagent
  gets its own context window and tool allowlist, so a deep investigation costs
  the main session a single summary line instead of ten thousand tokens.
- **[`pi-context-shake`](https://github.com/Integralist/pi-context-shake)**:
  An extension inspired by omp's `/shake` that shakes bloated content (massive
  tool output, images, verbose thinking) out of session history in place. It
  atomically rewrites the session's `.jsonl` file and uses a context hook to
  keep the live LLM payload trimmed, saving tokens and stopping annoying
  compaction loops without losing the running session.
- **[`model-stats`](https://github.com/Integralist/agent-skills/blob/main/.agents/skills/model-stats/SKILL.md)**:
  A skill that parses local logs across every harness I use (Claude Code, Pi,
  Codex, OpenCode, Gemini CLI, Copilot CLI) and renders an interactive HTML
  dashboard showing token consumption, spending, and reasoning effort by
  provider, model, and project. When a harness doesn't log costs directly, it
  estimates them using LiteLLM pricing data.
- **`caveman`**: An ultra-compressed communication mode (~75% token reduction)
  for rapid back-and-forth debugging when full conversational prose is just in
  the way.

## 🎬 Examples in Action

### 🚀 The BCP Workflow

The `bcp` skill ties together the final shipping loop: "Branch, Commit, Pull
Request".

A quick caveat before watching the recording below: performance here looks a
bit sluggish, but that is entirely down to my own laziness. I recorded this
using my session defaults (GPT 6.1 Sol with `xhigh` thinking effort), which is
massive overkill for mechanical git operations. Running this against a lighter
model like GPT Luna on `medium` effort makes the whole flow significantly
faster (and cheaper).

You might also wonder whether having an agent create branches, commit code, and
open PRs is complete overkill. Personally, I don't think so. Crafting
descriptive branch names, staging logical commits with useful context, and
writing PR descriptions that clearly outline the problem, solution, and
supporting diagrams takes genuine effort and mental bandwidth. Offloading that
ceremony lets me stay focused on the actual problem without cutting corners on
repository hygiene.

<script src="https://asciinema.org/a/Sn4shgXE3mYhLdgW.js" id="asciicast-Sn4shgXE3mYhLdgW" async="true"></script>

## 🏁 Wrapping Up

I don't think coding agents are going to replace engineers anytime soon.
Equally, writing them off as useless gimmicks misses how much time they can save
when you treat them properly.

If you let an agent write code without tests or boundaries, you'll end up with
unmaintainable junk. Wrap them in a sensible pipeline, make TDD non-negotiable,
and demand proof for every claim, and they become genuinely useful.

If you want to borrow any of the skills or conventions, have a browse around the
[`agent-skills`](https://github.com/Integralist/agent-skills) repo and tweak
them to suit your own workflow.
