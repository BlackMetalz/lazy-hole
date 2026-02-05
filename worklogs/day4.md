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

## Story 3.6: Add packet loss
Command: `sudo tc qdisc add dev eth0 root netem loss 10%`

Need to handle duplicate also
```bash
root@kienlt-jump:~# tc qdisc add dev eth0 root netem loss 5%
root@kienlt-jump:~# tc qdisc add dev eth0 root netem loss 5%
Error: Exclusivity flag on, cannot modify.
```

Hmm, rename `removeLatency` to `removeTCRules`

## Story 3.7: Block traffic

Command: `sudo iptables -A INPUT -s <IP> -j DROP`

And yeah, common cheatsheet for iptables here bro:
```bash
Iptables is Linux's classic firewall tool for managing packet filtering and NAT rules. Here's a cheatsheet of common commands for quick reference.

## Basic Commands
- List all rules: `iptables -L -v -n`
- List with line numbers: `iptables -L -v -n --line-numbers`
- Flush all rules: `iptables -F`
- Set default policy (e.g., DROP input): `iptables -P INPUT DROP`
- Save rules: `iptables-save > /etc/iptables.rules`
- Restore rules: `iptables-restore < /etc/iptables.rules` [gist.github](https://gist.github.com/davydany/0ad377f6de3c70056d2bd0f1549e1017)

## Common Rules
- Allow SSH (port 22): `iptables -A INPUT -p tcp --dport 22 -j ACCEPT`
- Allow HTTP/HTTPS: `iptables -A INPUT -p tcp --dport 80 -j ACCEPT` and `--dport 443`
- Allow established connections: `iptables -A INPUT -m state --state ESTABLISHED,RELATED -j ACCEPT`
- Block all other input: `iptables -A INPUT -j DROP`
- Allow loopback: `iptables -A INPUT -i lo -j ACCEPT` [andreafortuna](https://andreafortuna.org/2019/05/08/iptables-a-simple-cheatsheet/)

## Delete Rules
- Delete by line number: `iptables -D INPUT 5`
- Delete specific rule: `iptables -D INPUT -p tcp --dport 22 -j ACCEPT`
- Delete chain: `iptables -X customchain` [andreafortuna](https://andreafortuna.org/2019/05/08/iptables-a-simple-cheatsheet/)

## NAT Examples
- Masquerade for outbound: `iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE`
- Port forward: `iptables -t nat -A PREROUTING -p tcp --dport 8080 -j REDIRECT --to-port 80` [digitalocean](https://www.digitalocean.com/community/tutorials/iptables-essentials-common-firewall-rules-and-commands)

## Chain Management
| Chain   | Purpose                  |
|---------|--------------------------|
| INPUT   | Packets to this host    |
| OUTPUT  | Packets from this host  |
| FORWARD | Routed packets          |
| PREROUTING | Incoming before routing |
| POSTROUTING | Outgoing after routing |  [bashsenpai](https://bashsenpai.com/resources/cheatsheets/iptables)

Run as root or with sudo. For persistence, use tools like iptables-persistent on Debian-based systems.
```