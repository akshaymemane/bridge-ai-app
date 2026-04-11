# Claude Execution Playbook

This file is a companion to `BridgeAIChat_PRD.md` and `CLAUDE.md`.

It exists to give Claude Code sharper execution guidance without modifying Anthropic's `CLAUDE.md`.

## Read Order

When starting a session in this repo, ask Claude to read:

1. `BridgeAIChat_PRD.md`
2. `CLAUDE.md`
3. `CLAUDE_EXECUTION_PLAYBOOK.md`

## What Good Looks Like

Claude performs best here when it works in narrow vertical slices instead of trying to build the whole product in one pass.

The preferred sequence is:

1. Clarify architecture decisions that are required for the next slice.
2. Implement only one milestone at a time.
3. Run checks.
4. Review for drift against the PRD.
5. Move to the next slice.

## Default Scope Discipline

Unless explicitly requested, stay inside the version plan from the PRD:

- V1: single device, single tool, tmux integration
- V2: multi-tool routing
- V3: multi-device support and async jobs
- V4: orchestration

Do not quietly add V2-V4 behavior while building V1 foundations.

## Recommended Build Order

### Milestone 1

Repository scaffold:
- React + Vite + TypeScript frontend
- Go gateway
- Go bridge-agent
- shared message contract docs

### Milestone 2

Single-device happy path:
- one registered device
- one tool
- one chat thread
- one tmux-backed session
- live stream from CLI to UI

### Milestone 3

Operational hardening:
- reconnect handling
- device online or offline state
- missing tmux detection
- missing CLI detection
- clear error states

### Milestone 4

V2 expansion:
- multiple tools per device
- explicit routing via `/tool` or `@tool`
- default tool fallback

## Important Decisions To Lock Early

Before significant implementation, Claude should force clarity on:

- message schema between UI, gateway, and agent
- stream event types and termination semantics
- how chat IDs map to tmux session names
- where lightweight metadata is stored
- what "device registration" means in V1
- how file outputs are detected and surfaced

## Agent Usage

Three project subagents are available in `.claude/agents/`:

- `bridge-product-architect`
- `bridge-ui-engineer`
- `bridge-systems-engineer`

Suggested usage pattern:

- Use `bridge-product-architect` before broad implementation or when requirements feel fuzzy.
- Use `bridge-ui-engineer` for frontend slices and UX polish.
- Use `bridge-systems-engineer` for gateway, agent, tmux, and protocol work.

## Prompt Patterns

### For planning

```text
Read BridgeAIChat_PRD.md, CLAUDE.md, and CLAUDE_EXECUTION_PLAYBOOK.md.
Use the bridge-product-architect agent.
Turn the PRD into the next implementation milestone with:
1. assumptions
2. decisions that must be locked now
3. acceptance criteria
4. risks
Keep it scoped to V1 only.
```

### For frontend work

```text
Read BridgeAIChat_PRD.md, CLAUDE.md, and CLAUDE_EXECUTION_PLAYBOOK.md.
Use the bridge-ui-engineer agent.
Implement the smallest complete frontend slice for V1 chat UI.
Before coding, restate assumptions and list files to change.
Do not add multi-device or orchestration features.
```

### For backend work

```text
Read BridgeAIChat_PRD.md, CLAUDE.md, and CLAUDE_EXECUTION_PLAYBOOK.md.
Use the bridge-systems-engineer agent.
Implement the smallest complete Go backend slice for V1:
- single device
- single tool
- tmux-backed session
- streamed output
State the message contract before coding.
```

### For review

```text
Read BridgeAIChat_PRD.md and CLAUDE_EXECUTION_PLAYBOOK.md.
Use the bridge-product-architect agent.
Review the current implementation for PRD drift, hidden complexity, and missing acceptance criteria.
List findings by severity.
```

## Anti-Patterns

Avoid prompts like:

- "Build the whole app"
- "Implement the entire PRD"
- "Make it production ready" with no milestone boundary
- "Add all future features"

These usually cause unnecessary abstraction, incomplete wiring, or scope drift.

## Practical Advice

If Claude starts overbuilding:
- ask it to restate the current milestone
- tell it to remove anything beyond that milestone
- ask for acceptance criteria before more code

If Claude gets vague:
- ask for concrete contracts, file-level plan, and definition of done

If Claude gets stuck:
- switch to one of the specialized agents and narrow the task
