---
title: "How I Use AI"
date: 2026-08-17
description: "A deep dive into my agent-skills configuration: harness-agnostic architecture, strict operational rules, MCP servers, and repeatable skill pipelines."
tags: [ai, tooling]
---

Most conversations about using AI for software engineering boil down to two
extremes: breathless hype about autonomous agents replacing engineers by next
Tuesday, or cynical dismissals based on an LLM botching a regex query.

Neither perspective is useful.

I use AI coding agents every single day across multiple codebases. But I don't
use them as magical oracles, nor do I let them spray unvetted code into my
repositories. I treat an agent like a fast, capable, but occasionally careless
junior or mid-level engineer. That means strict boundaries, deterministic
workflows, mandatory test-driven development, and zero tolerance for sycophancy
or boilerplate slop.

All of this is codified in my
[`agent-skills`](https://github.com/Integralist/agent-skills) repository.
Here is a look under the hood at the mental model, architecture, and daily
workflows that make AI genuinely effective in my engineering work.

## The Mental Model: Harness Agnostic

The agent ecosystem is moving quickly. New CLI harnesses and models pop up
every few months: Claude Code, Pi, Gemini CLI, OpenCode, Copilot CLI, and
others.

Tying your workflows, prompt templates, and custom tools to one proprietary
agent harness is a fool's errand. The moment you switch tools, you lose all your
institutional habits.

My `agent-skills` repository solves this with a single source of truth:

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

The core directory is `.agents/skills/`. Every skill is written using generic,
harness-agnostic instructions ("prompt the user", "spawn a subagent", "read the
file"). Claude Code accesses them via a symlink
(`.claude/skills -> ../.agents/skills`), while other harnesses (like Pi, Gemini,
or OpenCode) load them directly from `~/.agents/skills/`.

When Claude Code supports unique capabilities (like path-scoped rules in
`.claude/rules/`), those rule files are automatically generated from the
corresponding canonical skills (`go-conventions`, `markdown-conventions`,
`sql-conventions`) via `make rules`.

```bash
# Install everything across all harnesses cleanly
make install
```

`make install` uses
[`scripts/op-inject.sh`](https://github.com/Integralist/agent-skills/blob/main/scripts/op-inject.sh)
to resolve 1Password secret references for API keys and MCP endpoints at
install time. If 1Password isn't authenticated, it gracefully skips those
injections rather than failing loudly.

## Rules of Engagement (`AGENTS.md`)

An agent without explicit behavioral constraints will default to people-pleasing
nonsense: apologizing profusely, hallucinating APIs, writing paragraphs of
fluff, and modifying files you never asked it to touch.

My global `AGENTS.md` sets the ground rules.

### 1. Communication and Tone

- **No sycophancy:** Cut out "Sure! I'd be happy to help with that!" Lead with
  the direct answer.
- **Conciseness:** Use the shortest complete response. Group lists by priority
  and cap them at roughly five items.
- **Bound multi-step work:** Number multi-step execution explicitly.

### 2. Engineering Discipline & TDD

- **Strict TDD:** No production code without a failing test first. Write the
  minimum code to pass, and delete assertions that survive an inverted
  requirement.
- **Simplicity over abstractions:** Solve problems by deleting components or
  reducing layers, not by stacking new frameworks or wrapper functions.
- **Cite sources:** Never rely on general memory for specific headers, API
  signatures, or configs. Cite the exact file and line number
  (`internal/parser/frontmatter.go:42`). If uncited, label it as an unverified
  assumption.

### 3. Skeptical Code Edits

- **Ask before editing:** A user question is an inquiry, not an open invitation
  to rewrite files. The agent must propose diffs in chat and get explicit
  approval before invoking code-editing tools.
- **Summarize large changes:** If a diff exceeds 40 lines, provide a one-line
  summary first and ask before dumping the whole diff or modifying the file.

## The Skill Pipeline: From Idea to Shipped PR

A common mistake is asking an agent to "build feature X" in a single prompt.
That almost always produces buggy, over-engineered slop.

Instead, I break work down into a pipeline of distinct, specialized skills.

```txt
clarify → grilling → architect → next-task → code-review → bcp
```

### Phase 1: Exploration and Stress Testing

Before any code is written, the concept must be validated.

- **`clarify`**: Elicits and pins down the user's core intent. If a request is
  vague, this skill asks targeted questions to eliminate ambiguity upfront.
- **`grilling`** (and `grill-me`): Puts the idea through an adversarial
  interrogation. The agent relentlessly challenges your assumptions, edge
  cases, and architecture choices across the design tree.
- **`perspectives` & `decide`**: Evaluates alternatives through structured
  lenses (risks, benefits, costs) and produces a durable decision memo or ADR
  (Architecture Decision Record).

### Phase 2: Architecture and Planning

Once the design survives the grilling phase, it gets structured.

- **`architect`**: Coordinates the transition from concept to concrete
  artifacts.
- **`to-spec`**: Generates a formal specification (`docs/specifications/`) with
  user stories, acceptance criteria, and testing seams.
- **`project-plan`**: Breaks the spec into vertical implementation slices with
  explicit `Blocked-by` dependency edges.
- **`tasks`**: Crystallizes the plan into a mechanical, TDD-shaped task list at
  `docs/tasks/` with exact code and test verification steps.

### Phase 3: Execution and Quality

With tasks written, execution is fast and deterministic.

- **`next-task`**: Reads the task file, executes the current step using TDD,
  verifies test output, and moves to the next item.
- **`go-conventions` / `go-testing`**: Enforces strict language conventions
  (table-driven tests, proper error wrapping, no unkeyed struct literals).
- **`cleanup`**: Runs a background audit specifically hunting for AI-generated
  clutter, dead code, and unnecessary abstractions.
- **`precedent`**: Audits newly written code against existing conventions in the
  codebase, flagging any inconsistencies in naming, error handling, or API
  signatures.

### Phase 4: Skeptical Review (Don't Be a Pushover)

Reviewing code with AI goes both ways: having agents review code, and critically
evaluating the feedback agents give you.

- **`code-review`**: Runs multi-dimensional reviews across parallel subagents
  (behavior, security, reliability, maintainability) to catch subtle defects.
- **`code-review-feedback` & `security-review-feedback`**: When an AI reviewer
  (or static analyzer) flags an issue, **do not reflexively accept it**. These
  skills force the agent to evaluate the claim with technical rigor. Is the
  vulnerability actually reachable? Is the suggested refactor introducing
  hidden complexity? If a suggestion is invalid, reject it with proof.

### Phase 5: Shipping

When the code is tested and clean, shipping is a single mechanical step.

- **`branch`**: Cuts a feature branch using session context and standard naming.
- **`commit`**: Stages and groups files intelligently with clean messages.
- **`draft-pr`**: Generates a concise pull request with clear Problem and
  Solution sections.
- **`bcp`**: Orchestrates branch creation, committing, and opening a PR in one
  command.

## Model Context Protocol (MCP) & Tooling

Agents are only as useful as the context they have access to. I use several MCP
servers to bridge the gap between local code and external tools:

- **Google Workspace MCP**: Bundled locally in `mcp/google-workspace/` (an
  unmodified build of upstream `gemini-cli-extensions/workspace`). Provides
  secure access to Calendar, Drive, Docs, Sheets, and Gmail via local OAuth.
- **Atlassian MCP**: Connects to Jira and Confluence using the modern MCP
  Streamable HTTP transport via `mcp-remote` with `--transport http-only`.
- **Language Servers (`gopls`)**: Provides real-time compiler diagnostics,
  symbol definitions, and type navigation directly to the agent.
- **Context7**: Live documentation indexing and retrieval for up-to-date
  third-party libraries.

> [!NOTE]
> All sensitive endpoints and API keys are injected dynamically via 1Password
> CLI templates (`.claude.json.tmpl`, `.copilot/mcp-config.json.tmpl`), ensuring
> no secrets ever leak into version control.

## Multi-Agent Orchestration & Cost Control

Running every simple command through top-tier frontier models is slow and
expensive. My setup enforces cost and delegation discipline:

- **Model Tiering**: Mechanical tasks (searching files, drafting changelogs,
  running tests) default to fast, cheap models (such as Gemini Flash or Claude
  Haiku). High-reasoning models are reserved for architecture, complex
  debugging, and grilling.
- **`pi-intercom`**: Allows multiple Pi agent sessions on the same machine to
  communicate, delegate subtasks, and share context in real time.
- **`pi-btw`**: Enables lightweight side-conversations without derailing the
  main agent thread or polluting the primary context window.
- **`caveman`**: An ultra-compressed communication mode (~75% token reduction)
  for rapid back-and-forth debugging when full conversational prose is just in
  the way.

## Conclusion

AI coding agents are neither replacing software engineers nor are they useless
gimmicks. They are powerful multipliers when paired with sound engineering
practices.

If you let an agent write code without tests, without planning, and without
architectural constraints, you will get unmaintainable junk. But if you wrap
the agent in a structured pipeline of modular skills, enforce strict TDD, and
demand source verification, it becomes one of the most effective tools in your
developer toolkit.

Feel free to explore the
[agent-skills repository](https://github.com/Integralist/agent-skills) and adapt
the skills and conventions for your own workflow.
