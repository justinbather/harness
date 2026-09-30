# harness

A terminal UI for looking inside a Kafka cluster. Point it at a broker and it consumes every
topic it can find, then lets you browse topics and inspect individual messages without leaving
the shell.

Built because a lot of the events my services emitted at work went off into other systems as a
black box, and I wanted to actually see them.

## Install

```bash
go install github.com/justinbather/harness@latest
```

## Usage

```bash
harness                        # defaults to localhost:9092
harness kafka.internal:9092    # or pass a broker
```

Two screens. The first lists every topic with its partition count and how many messages
`harness` has consumed. Press enter to drop into a topic and page through its messages by
partition and offset.

### Keys

| Key | Action |
| --- | --- |
| `j` / `k` | Down / up |
| `d` / `u` | Half page down / up |
| `enter` | Open the selected topic |
| `y` | Copy the selected message payload to the clipboard |
| `esc` / `ctrl+o` | Back to the topic list |
| `q` / `ctrl+c` | Quit |

## How it works

`harness` discovers topics through a metadata request, then starts a consumer across all of
them and writes everything it receives into an in-memory store. The TUI reads from that store.
Nothing is written to disk and no offsets are committed, so it's safe to point at a cluster
without disturbing existing consumer groups — though it does read every topic it finds.

| Package | Role |
| --- | --- |
| `internal/harness` | Kafka client lifecycle, topic discovery, the consume loop. |
| `internal/store` | Mutex-guarded in-memory topic and message storage. |
| `internal/logger` | Context-scoped `zap` logger. |
| `main.go` | Bubble Tea model, the two screens, keybindings. |

Built on [franz-go](https://github.com/twmb/franz-go) for Kafka and
[Bubble Tea](https://github.com/charmbracelet/bubbletea) for the TUI.

## Known limitations

- **Live refresh is written but switched off.** `poll()` ticks every two seconds and the wiring
  exists, but re-rendering the topic table resets cursor position because the rows come out of a
  map in non-deterministic order. Enabling it needs a stable sort and cursor preservation first.
- **The store is read without holding its mutex.** `ListTopics` and `ListMessages` return the
  live map and slice while the consume goroutine is appending to them.
- **Everything is buffered in memory** with no cap, so a busy cluster will grow the process
  without bound.
- Errors inside the update loop `panic` rather than surfacing in the UI.
- Consumes from the beginning of every topic it finds, which is fine locally and inadvisable
  against a large production cluster.
