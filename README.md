# castctl

I got tired of clicking around the GCP console to babysit live streams, and
`gcloud` doesn't really cover the Live Stream API, so this is a small CLI that
does. It talks to both the Live Stream API and the Transcoder API, ships as one
binary, and uses whatever Application Default Credentials you already have.

Windows, mac, Linux. No runtime, no python, nothing to `pip install`. Just the binary.

## Getting it

Grab a build off the releases page, or build it yourself.

On Windows there's a script that builds it and drops it on your PATH:

```powershell
powershell -ExecutionPolicy Bypass -File install.ps1
```

It lands in `%LOCALAPPDATA%\Programs\castctl`. You'll need to open a fresh
terminal after, PATH changes don't apply to the window you ran it from.

mac / Linux:

```bash
make install      # goes to ~/.local/bin, make sure that's on your PATH
```

Or if you have Go and don't care where it ends up:

```bash
go install github.com/simeon/castctl@latest
```

## Auth

It's ADC, so however you'd auth anything else against GCP works here. The usual:

```bash
gcloud auth application-default login
```

For CI or a box with no browser, point it at a service account key instead:

```bash
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/key.json
```

Not sure if it's picking creds up? Ask it:

```bash
castctl auth status
```

If creds are missing it'll tell you exactly what to run, so you don't have to
go digging.

## Project and location

Every command needs a project and a region. You can pass them each time with
`--project` / `--location`, but that gets old fast. Set them once:

```bash
castctl config set project my-project
castctl config set location us-central1
```

Lives in `~/.castctl/config.yaml`. A flag always wins over the config file, and
the env vars (`GOOGLE_CLOUD_PROJECT`, `CASTCTL_LOCATION`) sit in between. So
flag > env > file, if you like it spelled out.

## What it can do

Commands are noun-then-verb, same shape as gcloud and kubectl, so muscle memory
carries over.

Live Stream side:

- `input`: list, get, create, update, delete
- `channel`: list, get, create, update, delete, and start / stop
- `event`: list, get, create, delete (pass `--channel`)
- `clip`: list, get, create, delete (pass `--channel`)
- `asset`: list, get, create, delete

Transcoder side:

- `job`: list, get, create, delete
- `job-template`: list, get, create, delete

## Output

You get a table by default because that's what you actually want when you're
just looking. When you need to pipe it somewhere, add `--json` and you get
proper protojson back:

```bash
castctl channel list
castctl channel list --json | jq '.[].name'
```

Handy bit: the JSON that comes out of a `get` is valid input for a `create`.
So `get` one thing, tweak the file, `create` the next. No format juggling.

## Creating stuff

Channels and jobs have big nested configs (elementary streams, mux streams, the
whole thing) and cramming that into flags would be miserable. So `create`
takes a JSON file with `-f` (or `-` to read stdin):

```bash
castctl input create my-input --type RTMP_PUSH
castctl channel create my-channel -f examples/channel-hls.json
castctl channel start my-channel
castctl job create -f examples/job-preset.json
```

Inputs are simple enough that `--type` alone works, but everything else wants a
spec. There's a few starters in `examples/` to copy from. One gotcha in
`channel-hls.json`: the input attachment's `input` field has to be the full
resource name (`projects/P/locations/L/inputs/my-input`), not just the short id.

## A note on waiting

Most Live Stream operations are async on Google's end. Create a channel, start
it, delete it, and you get back an operation that finishes a little later. By
default castctl just waits for it and doesn't come back til it's done. If you'd
rather fire and forget, `--async` hands you the operation name and returns
immediatly. Transcoder jobs are synchronous so this doesn't apply there.

## Tab completion

There's completion for bash, zsh, fish and powershell. Pick yours:

```powershell
# powershell, one-off for this session:
castctl completion powershell | Out-String | Invoke-Expression
# or make it stick:
castctl completion powershell >> $PROFILE
```

```bash
# bash
echo 'source <(castctl completion bash)' >> ~/.bashrc

# zsh
castctl completion zsh > "${fpath[1]}/_castctl"

# fish
castctl completion fish > ~/.config/fish/completions/castctl.fish
```

## Building for everything at once

```bash
make dist
```

That spits out windows/amd64, darwin arm64 + amd64, and linux/amd64 into
`dist/`. It's all pure Go with `CGO_ENABLED=0`, so cross-compiling is nothing
fancy, no toolchains to install. Pushing a `vX.Y.Z` tag runs the same builds
in CI and attaches the binaries to the release.

## Missing bits

DVR sessions and event pools aren't wired up yet, didn't need them. If you do,
the pattern's the same as the other resources, so it's not much work to add.
PRs welcome, or just open an issue and tell me what's broken.
