<!-- Generated from the canonical exercise source (docs/01-message-pump/SOLUTION.md) by emit.py, which lives in the practical-messaging-samples working directory alongside the five language repos. Do not edit this copy -- edit the canonical source and re-emit. -->
# Exercise 1 — what the fix is #

**This file is prose on purpose.** There is no code in it.

If you have an agent, this tells you whether you asked for the right thing. If you do not have
an agent, this tells you what to type. Either way the interesting part of the exercise already
happened, in the blanks in `PROBE.md` — **the answer is not the point; the prediction was.**

A complete, compiling version of all of this is sitting in `../02-failing-well/`, which is
where exercise 2 starts from. Read that if you want the code.

---

## The three defects, and why each one is a defect ##

### 1. The handler took a `amqp091.Delivery` ###

`PlaceOrderHandler.Handle` received the raw AMQP delivery and deserialized the body itself. So
the pump had a **Get** stage and a **Handle** stage and no **Translate** stage at all — the
translation had leaked into the application code.

What it costs you, in the order the costs actually arrive:

- **`model/placeorderhandler.go` had to import `github.com/rabbitmq/amqp091-go`.** The domain now has a
  compile-time dependency on a specific broker: build `model/` and you build the
  broker's client library, and a test of the handler links it too. Swap broker and the domain
  does not compile.
- **The handler cannot be called by anything else.** Not a unit test, not an HTTP endpoint, not
  a developer poking at it from a console app — every caller has to fabricate a
  `amqp091.Delivery`, which means fabricating a delivery tag and a byte array.
- **The pump cannot tell your failure from its own, and in Go it cannot even see that there
  are two of them.** A body the handler could not decode and a SKU that is not in the
  catalogue come back from the same call as an `error` — one interface, one method, no type
  to ask about. The pump has no way to know which is "I could not understand this" and which
  is "I understood it and the work failed". That distinction is the whole of exercise 2, and
  this signature makes it unavailable.

**The fix:** a `Message Mapper` — a function or class whose job is body → `PlaceOrder`. The pump
calls it. The handler takes `PlaceOrder` and returns. Delete the `import` from
`model/placeorderhandler.go` and let the compiler find anything you missed.

▎ The check is mechanical and it is the one from the deck: **if a handler's signature has a broker type in it, the mapper has not finished its job.**

### 2. The pump acknowledged before handling ###

```
    delivery, ok, err := consumer.Receive()
    consumer.Acknowledge(delivery.DeliveryTag)   // <-- here
    err = p.handler.Handle(delivery)
```

The comment above it said *"we have the message in our hands, so the broker does not need to
hold it for us any more."* That sentence is wrong in one word: we have the **message**, but we
have not done the **work**.

An ack means *I am done with this; you may forget it*. Sending it before the work is done tells
the broker a lie, and the broker believes you — which is what Probe C showed. `unacked=0`
during thirty seconds of work is the broker reporting, accurately, that it has nothing
outstanding. It has nothing outstanding because you told it so.

**The fix:** acknowledge **after** the handler returns successfully. Nothing else moves.

**What this costs, and it is worth saying out loud:** you have just chosen **at-least-once**.
If the process dies between the handler finishing and the ack landing, the order is placed and
the message comes back, and it will be handled twice. That is not a bug you can fix by moving
the ack — move it back and you get at-most-once and silent loss instead. **There is no third
position.** The answer to the duplicate is an *Inbox*, on the consumer side, and it is §4.4's
last slide rather than this exercise.

### 3. A failure in any stage killed the pump ###

There is no `try` and no `catch` here, because Go has neither — and it makes no difference.
The mapper and the handler hand their failure back as an `error`, and the loop's only answer
to an `error` is `return err`. So the failure unwinds out of `Run`, out of `main`, which
passes anything `Run` gives it to `log.Fatalln`, and takes the process with it. One malformed
body from a producer you do not control is a denial of service on your consumer.

▎ **Nothing was thrown and the outcome is identical**, which is the thing to notice: this defect was never about exceptions. It is about a loop with no answer to a failure except to stop being a loop.

**The fix, for now, is deliberately partial.** Keep the error from the translate-and-dispatch
part of the loop instead of returning it, do **not** ack a message you failed to handle, and
log enough to know it happened. That is enough to make the pump stop dying.

It is **not** enough to be correct, and you should be able to say why:

- If you `Reject(tag, true)` an unmappable body, you have built an infinite loop. It will
  never map. It will come straight back. Your log will fill up at the speed of the network.
- If you `Reject(tag, false)` it, the broker deletes it — the message is still gone, you
  have just stopped dying while losing it.
- If you ack it to get rid of it, you have silently discarded a message that a misconfigured
  producer may have sent in perfectly good faith, and nobody will ever know.

**All three are wrong, and they are wrong in the direction exercise 2 fixes.** The message needs
somewhere to *go*: an **Invalid Message Channel** for a body you could not read, and a **Dead
Letter Channel** for work you retried and gave up on. Those are two different places because
they are two different failures, and RabbitMQ will give you both without you writing a queue.

---

## What a correct exercise 1 looks like, in four sentences ##

The gateway owns the broker; nothing outside `simplemessaging` names a RabbitMQ type. The pump
does Get, then Translate through a mapper, then Dispatch to a handler, then acknowledge — in
that order, and the acknowledge is last. A failure in any stage is caught, so the loop survives
to take the next message. And the handler is an ordinary method over a domain type that would
run identically if you called it from a test.

▎ Your handler is not a message handler. It is a method that happens to be called by one.
