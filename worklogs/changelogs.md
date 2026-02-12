# Issue Feb 12, 2026
bug found: when i added 6 ip for blackhole. in 3 hosts. But when i quit

```
Cleaning up effects...
Warning: failed to remove blackhole on mysql-galera-3: Route not exists: 192.168.13.13
Warning: failed to remove blackhole on mysql-galera-3: Route not exists: 192.168.13.13
Warning: failed to remove blackhole on mysql-galera-1: Route not exists: 192.168.13.13
Warning: failed to remove blackhole on mysql-galera-1: Route not exists: 192.168.13.13
Warning: failed to remove blackhole on mysql-galera-2: Route not exists: 192.168.13.13
Warning: failed to remove blackhole on mysql-galera-2: Route not exists: 192.168.13.13
Restored 3 hosts
```

Reality in mysql-galera-3
```bash
ip route|grep black
blackhole 192.168.3.12
blackhole 192.168.13.11
```

So it failed to clean all effects.

Fix should be acceptable event i will copy some slice, but it is the easiest way at this fucking time! Holidays is comming!!!