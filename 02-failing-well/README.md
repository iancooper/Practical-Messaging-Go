<!-- Generated from the canonical exercise source (docs/02-failing-well/README.md) by emit.py, which lives in the practical-messaging-samples working directory alongside the five language repos. Do not edit this copy -- edit the canonical source and re-emit. -->
# Exercise 2 — Failing Well #

**Start with [PROBE.md](PROBE.md).** 40 minutes.

```
go run ./cmd/receiver  # the pump
go run ./cmd/sender    # a good order
go run ./cmd/sender unmappable | poison | flaky | slow | burst 20
```

Management console: <http://localhost:15672> — filter Queues on `failing-well`. There are four.

| | |
|---|---|
| [`PROBE.md`](PROBE.md) | the exercise |
| [`SOLUTION.md`](SOLUTION.md) | what the fix is, in prose. After the predictions, not before |
| `simplemessaging/channel.go` | the topology, drawn in a comment. Read it |
| `simplemessaging/messagepump.go` | **the one file you need to change** |

**This is exercise 1's answer**, so you start level whether or not you finished it. The pump no
longer crashes and no longer loses messages — and it is still wrong.
