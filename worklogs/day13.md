# Day 12: Mar 02, 2026
# Trying to finish some backlog with help of AI xD

### Group view missing active rule applied
More issue appear while trying to implement this shit

- Undo action on group not clean view on effected hosts. It still show active rule on hosts. --> probably need to rebuild layout...

- Not really, that is how it works, in group we block a group, that group have 3 fucking hosts, need to undo 3 times? pretty bad for UX, need to ~~fix~~ promt xD

- Third, when I exit TUI, it still show 0 active rule on hosts. --> need to correction. Count total effects instead of check `effectTracker.GetAll()` > 0