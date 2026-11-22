# Grimhold

Grimhold is a text-based dungeon game that runs in the terminal, written in Go with only the standard library. The player types commands in a loop (REPL) to travel between cursed regions described in `world.json`. They explore, stumble on monsters, fight them with dice, loot weapons from them, then record their victims in a bestiary. Monster and weapon data is fetched live from the [Open5e](https://open5e.com) API (D&D 5e SRD, free, no key), with an in-memory cache. A game can be saved to `save.json` and resumed later.

## Requirements

- Go 1.27+
- Make
- golangci-lint (optional)

## Quickstart

```sh
git clone https://github.com/Sol1de/go-grimhold.git
cd go-grimhold
make build
./bin/grimhold
```

Or run without building:
```sh
make run
```

### Playing

At the `grimhold >` prompt, type `help` for the full command list. A typical first session:

```text
look                  # describe where you are
map                   # regions, danger level and exits
travel ashen-marsh    # move to a neighboring region
explore               # search for a monster
fight                 # attack it (or: flee)
loot                  # after a victory: find a random weapon
equip greatsword      # wield a weapon from your bag
character             # level, weapon and bag
status                # HP, XP, gold, kills
rest                  # full heal, once per region visit
exit
```

### Weapons

You start with a Longsword (1d8). After each victory, `loot` adds a random weapon from Open5e to your bag. `equip <weapon>` wields it: your hits then deal its damage dice +2.

### Ambushes

While you are in a region with monsters and not fighting, a monster may ambush you without you typing anything. Every 20 seconds there is a danger-in-10 chance (10% in a danger 1 region, 30% in danger 3). An ambush starts a fight, just like `explore`. There are no ambushes during a fight or at the crossroads.

### Saving

- `save` writes the game to `save.json` in the current directory (not allowed mid-fight).
- `load` restores it during a session.
- `./bin/grimhold -continue` loads `save.json` on startup.

### Development

```sh
make run    # go run ./cmd/grimhold
make test   # go test ./...
make vet    # go vet ./...
make fmt    # go fmt ./...
make lint   # requires golangci-lint
make clean  # remove bin/
```
