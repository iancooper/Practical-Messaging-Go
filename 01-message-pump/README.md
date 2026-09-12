<!-- Generated from the canonical exercise source (docs/01-message-pump/README.md) by emit.py, which lives in the practical-messaging-samples working directory alongside the five language repos. Do not edit this copy -- edit the canonical source and re-emit. -->
# Exercise 1 — The Message Pump #

**Start with [PROBE.md](PROBE.md).** It is the exercise. 35 minutes.

```
go run ./cmd/receiver  # the pump
go run ./cmd/sender    # sends one order
```

Management console: <http://localhost:15672> (`guest` / `guest`) — keep it open.

| | |
|---|---|
| [`PROBE.md`](PROBE.md) | the exercise: read, predict, break, fix |
| [`SOLUTION.md`](SOLUTION.md) | what the fix is, in prose. Read it after the predictions, not before |
| `simplemessaging/` | the messaging gateway. Given to you, and correct |
| `model/` | the domain — an order, a catalogue, and a handler |
| `cmd/sender/`, `cmd/receiver/` | two console apps |

**The pump in `simplemessaging/messagepump.go` is deliberately wrong.** It runs. Do not copy it.
