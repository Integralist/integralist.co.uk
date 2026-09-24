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
highlighting, and side conversations. So I built extensions for them:
[`pi-statusbar`](https://github.com/Integralist/pi-statusbar),
[`pi-subagents`](https://github.com/Integralist/pi-subagents),
[`pi-btw`](https://github.com/Integralist/pi-btw), and
[`pi-transcript-enhancer`](https://github.com/Integralist/pi-transcript-enhancer).

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

When Claude Code supports handy features like path-scoped rules in
`.claude/rules/`, I generate them automatically from the canonical conventions
skills (`conventions-go`, `conventions-python`, and so on) via `make rules`.

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
                                                       ├─→ to-spec → to-plan → to-tasks → next-slice
Engineering track: architect (research → grill) ───────┘
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
2. **Discovery (`discovery.md`)**: Evaluates solution directions and trade-offs
   for a milestone without committing to implementation details.
3. **Design (`design.md`)**: Documents the approved system-level architecture
   thoroughly enough for Product, Engineering, and reviewers to sign off.

Unlike most agent skills that run through to the end autonomously, `pdd` stops
dead at each gate (`Awaiting approval`). It refuses to advance to Discovery
until Project is approved, and will not touch Design until Discovery is signed
off.

Crucially, `design.md` is not an implementation plan (`plan.md`). It answers
what system solution is being approved, not how an engineer will slice the pull
requests. Once Design is approved, `pdd` hands off directly to `to-spec` to
define the technical contracts. You do not run `architect` afterwards, because
doing so would pointlessly duplicate the discovery you just agreed upon.

#### Engineering-led coordination: `architect`

When an initiative is driven entirely by engineering, running through a
three-stage product governance dance is overkill. That is where `architect`
comes in.

Where `pdd` is a gated approval workflow for cross-functional consensus,
`architect` is an automated coordinator for engineering delivery. It chains
five phases together in sequence: bootstrapping project rules (`agents-md`),
deep research (`research`), writing the functional specification (`to-spec`),
adversarial stress-testing (`grill-with-docs`), and implementation planning
(`to-plan`).

If `pdd` is about deciding what to build with Product, `architect` is about
taking an engineering idea and autonomously generating the technical artefacts
needed to build it safely.

#### The engineering handoff: specs, plans, and tasks

Regardless of whether an initiative starts in `pdd` or `architect`, both tracks
converge on the same engineering delivery pipeline under
`projects/<yyyy-mm-dd-slug>/`:

- **`to-spec`**: Generates a solid capability specification (`spec.md`) with
  user stories, acceptance criteria, and testing seams. Living specs stay
  validated in CI so contracts don't rot.
- **`to-plan`**: Breaks the specification into vertical implementation slices
  (`plan.md`) with explicit `Blocked-by` dependency edges and PR groupings.
- **`to-tasks`**: Compiles a single plan slice just-in-time into a mechanical,
  TDD-shaped runbook (`tasks.md`) with verbatim code and assertions, avoiding
  upfront plan decay.

### 🔨 Phase 3: Building (Small, Mechanical Slices)

With tasks compiled, execution becomes fast and predictable.

- **`next-task` & `next-slice`**: `next-task` implements and verifies the next
  item in the runbook, while `next-slice` works through an entire vertical plan
  slice in one go.
- **`conventions-*` & `go-testing`**: Enforces strict language conventions
  (`conventions-go`, `conventions-python`, `conventions-sql`, and others) and
  testing discipline (table-driven tests, `t.Context()`, bounded waits, public
  API testing, and zero unkeyed struct literals).
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

- **`code-review`**: Runs multi-dimensional reviews across parallel subagents
  (behaviour, security, reliability, maintainability) using `pi-subagents` to
  catch subtle bugs without blowing up the context window.
- **`code-review-feedback` & `security-review-feedback`**: When an AI reviewer
  (or static analyser) flags an issue, you shouldn't reflexively accept it.
  These skills force the agent to evaluate the claim with technical rigour. Is
  the vulnerability actually reachable? Is the suggested refactor introducing
  hidden complexity? If a suggestion is rubbish, reject it with proof.

### 🚢 Phase 5: Shipping Without the Drama

When the code is tested and clean, shipping is a deterministic, mechanical step.

- **`branch`**: Cuts a feature branch using session context and standard naming.
- **`commit`**: Stages and groups files sensibly, enforcing titles that explain
  why the change matters rather than listing raw mechanical steps.
- **`draft-pr`**: Generates a concise pull request with clear Problem and
  Solution sections, consequence-driven titles, and load-bearing emoji.
- **`stacked-prs`**: Coordinates dependent PR chains with the official
  `gh-stack` CLI extension, mapping vertical slices from `to-plan` directly into
  isolated, easily reviewable PR layers.
- **`bcp`**: Orchestrates branch creation, committing, and opening a PR (or
  submitting a stacked branch) in one command.

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

## 🔍 Tight Feedback Loops with Crit

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
pushing a single commit upstream. And when everything finally looks right, `crit
push` can sync those local comments straight up to the GitHub PR review if
needed. It keeps the human firmly in the driving seat without breaking the flow.

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

> [!NOTE]
> All sensitive endpoints and API keys are injected dynamically via 1Password
> CLI templates (`.claude.json.tmpl`, `.copilot/mcp-config.json.tmpl`), ensuring
> no secrets ever leak into version control.

## 💰 Keeping Costs Sane

Running every simple command through top-tier frontier models is slow and
expensive. My setup enforces a bit of financial and token discipline:

- **Model Tiering**: Mechanical tasks (searching files, drafting changelogs,
  running tests) default to fast, cheap models like Gemini Flash or Claude
  Haiku. Heavy reasoning models are kept for architecture, gnarly debugging, and
  grilling.
- **[`pi-subagents`](https://github.com/Integralist/pi-subagents)**: An
  extension I built to run isolated subagents in the same process. Each subagent
  gets its own context window and tool allowlist, so a deep investigation costs
  the main session a single summary line instead of ten thousand tokens.
- **[`model-stats`](https://github.com/Integralist/agent-skills/blob/main/.agents/skills/model-stats/SKILL.md)**:
  A skill that parses local logs across every harness I use (Claude Code, Pi,
  Codex, OpenCode, Gemini CLI, Copilot CLI) and renders an interactive HTML
  dashboard showing token consumption, spending, and reasoning effort by
  provider, model, and project. When a harness doesn't log costs directly, it
  estimates them using LiteLLM pricing data.
- **`caveman`**: An ultra-compressed communication mode (~75% token reduction)
  for rapid back-and-forth debugging when full conversational prose is just in
  the way.

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
