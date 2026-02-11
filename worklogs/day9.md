# Day 9: Feb 11, 2026

## Story 6.1: Refresh Host Status
Goal: When user press `r` in host list, refresh host status!

Test case: Add sudo for specific host without sudo, refresh to see host getting update or not.

Refresh host can lead to order of host changed, i wonder i should make it persistent like order or config file or not?


Result Before:
![alt text](../images/02.png)

After hit button refresh:
![alt text](../images/12.png)

And yeah, there is no guide for user know where the fuck is button r do. So we have to put that shit to 6.2