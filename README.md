# delaystat

`delaystat` is a cli utility that fetches and prints linux delay accounting information for a given process or thread.

### Usage

Delay accounting for by PID:

```bash
$ delaystat -p 1
```

Delay accounting by TGID:

```bash
$ delaystat -t 1
```

If no flags are provided, `delaystat` prints its own delay accounting information.

### More information

For more information on linux delay accounting, please see:

- https://www.kernel.org/doc/Documentation/accounting/delay-accounting.txt
- https://andrestc.com/post/linux-delay-accounting/

### Notes

 - Requires CAP_NET_ADMIN capability (see capabilities(7)). Otherwise, the application must be run as root.
