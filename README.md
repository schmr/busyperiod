# busyperiod
> Random task set EDF busy period check

This PoC includes a random task set generator,
the schedulability check of EDF-NUVD,
and a busy period check for EDF scheduled task sets.

It generates a lot of random task sets, checks them for EDF-NUVD schedulability,
and applies the EDF busy period check to investigate the low criticality mode
behaviour of the task set in the worst case.

If the busy period check fails, a virtual deadline is missed in low criticality
mode. This counterexample is saved, and the execution stops.
The whole application can be interpreted as a brute force search for a
counterexample.


## License

[MIT © Robert Schmidt](./LICENSE)
