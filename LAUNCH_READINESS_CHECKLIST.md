# Launch Readiness Checklist

Legend:

- `Done` = already validated
- `Partial` = exists, but not strong enough for a broad launch
- `Pending` = still needs work

## Core Product

- `Done` Gateway exists and runs.
- `Done` Frontend chat UI exists and runs.
- `Done` Standalone agent repo exists and was pushed.
- `Done` Standalone docs repo exists and was pushed.
- `Done` Device discovery from Tailscale works.
- `Done` Device state merge works: `connected`, `agent_missing`, `offline`.
- `Done` At least one real agent connection was verified end to end.
- `Partial` Chat-to-AI transport works, but the agent still feels more like a connector than an intelligent remote assistant.
- `Pending` Tool/session switching feels product-grade and predictable.

## First-Run User Experience

- `Done` There is a simple beta login flow via tailnet input.
- `Partial` The setup flow is understandable for technical users.
- `Pending` A brand-new user can install and succeed without your help.
- `Pending` One clear golden path from download to first successful reply has been tested from scratch after the repo split.
- `Pending` Mobile-first usability is validated, which was part of the original reason for building this.

## Agent Quality

- `Done` Agent can connect and register against the gateway.
- `Done` Agent can expose available tools.
- `Partial` Agent can relay prompts to local AI CLIs.
- `Pending` Agent has lightweight built-in intelligence for remote assistance.
- `Pending` Agent behavior feels trustworthy enough that a user would prefer it over just SSH.
- `Pending` Clear handling for common remote tasks like logs, files, process checks, and simple machine actions.

## Reliability

- `Done` App repo builds passed.
- `Done` Agent repo build passed before split cleanup.
- `Done` Docs repo build passed before split cleanup.
- `Partial` Core happy-path runtime works locally.
- `Pending` Fresh-machine install verification for app release archive.
- `Pending` Fresh-machine install verification for agent release archive.
- `Pending` Multi-device testing with more than one real active agent in normal usage.
- `Pending` Reconnect and recovery behavior is hardened enough for broad users.
- `Pending` Broader failure handling and recovery UX inside the app.

## Documentation

- `Done` Docs website exists as a standalone repo.
- `Done` Docs were updated to reflect the current beta story.
- `Partial` Docs are probably good enough for technical testers.
- `Pending` Docs have been followed by an external person with no help and proven complete.
- `Pending` Install docs exactly match the final published download and install path after the repo split.

## Distribution

- `Done` Repos are cleanly split into app, agent, and docs.
- `Done` All three repos were pushed.
- `Partial` GitHub-release-based distribution story exists.
- `Pending` Final release artifacts are produced and verified from the standalone repos after split.
- `Pending` Download and install page is validated against real published release assets.

## Trust and Launch Risk

- `Done` The idea is real and demonstrable.
- `Done` There is enough working product for a private or technical beta.
- `Pending` The product is smooth enough for a broad public come-use-it launch.
- `Pending` A few outside testers have succeeded without direct support.
- `Pending` You are confident first impressions will be useful rather than interesting but rough.

## Bottom Line

You are `Done` on the foundation and `Partial` on product readiness.

For a broad public launch, the biggest pending items are:

- product-grade onboarding
- stronger agent behavior
- predictable tool and session UX
- fresh-install validation from published artifacts
- external tester proof
