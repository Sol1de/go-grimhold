# Grimhold

Grimhold is a text-based dungeon game that runs in the terminal, written in Go with only the standard library. The player types commands in a loop (REPL) to travel between cursed regions described in `world.json`. They explore, stumble on monsters, fight them with dice, then record their victims in a bestiary. Monster data is fetched live from the [Open5e](https://open5e.com) API (D&D 5e SRD, free, no key), with an in-memory cache. A game can be saved to `save.json` and resumed later.

## Quickstart

TODO
