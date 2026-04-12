# bridge-homebrew-tap

This directory is the seed for the future standalone `bridge-homebrew-tap` repository.

It is the public Homebrew tap for:

- `bridge-app`
- `bridge-agent`

## Intended user install

```bash
brew install bridge-ai-chat/tap/bridge-app
brew install bridge-ai-chat/tap/bridge-agent
```

## Repo layout

- `Formula/bridge-app.rb`
- `Formula/bridge-agent.rb`

## Update flow

1. Publish a new tagged release in `bridge-app`
2. Publish a matching tagged release in `bridge-agent`
3. Download the rendered formula artifacts from both release workflows
4. Copy them into `Formula/`
5. Commit and push this tap repo

## Notes

- The formula files should always point at GitHub Release assets.
- The first beta keeps GitHub Releases as the source of truth for downloads.
- Formula generation happens in the app and agent repos so checksums stay aligned with the release artifacts.
