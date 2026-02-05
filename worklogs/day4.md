# Day 4: Feb 5, 2026

## Story 3.5: Remove latency 

Command: `sudo tc qdisc del dev eth0 root`

Handle when delete rule not exists
```bash
root@kienlt-jump:~# tc qdisc del dev eth0 root 2> abc.txt
root@kienlt-jump:~# cat abc.txt
Error: Cannot delete qdisc with handle of zero.
```

So you clearly can see we direct error to `abc.txt`. That is why we use to check strings contains in `result.Stderr`

