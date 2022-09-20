
## Done() & Flush()

We want to call Done() on base spans as soon as we know they're
done, or just before a Flush() if there is activity since the last
Done().

Each log keeps track of it's non-detached children that aren't Done.

A log can only be Done() if it's dependent children are Done(). 

activeChildren are those that are not explicitly Done() (both dependent and detached)

Non-detached children are Done when their parent is Done or if Done() is called
on them explicitly.

span.knownActive: true if listed in parents activeChildren

## Ordering

Change atomics and then take action based upon the value of the 
atomic.  For example set knownActive to 1 and then if it had been
zero, add to depndentents and such.

On the reverse, set knownActive to 0 and then call Done(), etc.

Try not to hold locks while calling things: make todo lists while
holding the lock but then do the work w/o the lock.  

## SpanModifiers and randomizing trace id / span id

Generally speaking, we randomize span ids early.

The usual sequence is 

1. `Seed`
2. `Request()` -> `log`
2. `log.Sub().Fork().`

XXX

