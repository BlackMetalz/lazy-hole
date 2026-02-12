# Day 10: Feb 12, 2026

I realized input single ip into blackhole is just waste of time, i have to do it 6 times for 3 hosts (So total 18 times for input!). So i decided to add feature to support multiple ip for blackhole!

Example input for blackhole: 192.168.3.21,192.168.3.22,192.168.3.23

Logic is pretty simple to understand, right?

![alt text](../images/18.png)

![alt text](../images/19.png)

Before and after
```
# Before
root@kienlt-jump:~# ip r|grep black
blackhole 10.0.0.1
blackhole 10.0.0.2
blackhole 10.0.0.3
# After
root@kienlt-jump:~# ip r|grep black
```

Exists test
![alt text](../images/20.png)

Ready to rock. Haha
So shit load of issue when we use it, not only for blackhole but also for others. I will fix them one by one!

And yeah, work with this shit give me a lot of questions. What is fucking different between blackhole and iptables block?
- blackhole: layer 3, drop all traffic to specific IP/CIDR
- iptables: layer 4, drop all traffic to specific port from specific IP

So literally same purpose but different layer! Haha.

So I need to make same feature for network partition like blackhole, support multi IP input. This is pretty simple to understand so I will ask for vibe code and I will review generated code!