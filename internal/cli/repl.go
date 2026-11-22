package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"sort"
	"strings"

	"github.com/Sol1de/go-grimhold/internal/client"
	"github.com/Sol1de/go-grimhold/internal/game"
	"github.com/Sol1de/go-grimhold/internal/world"
)

type cliCommand struct {
	name        string
	description string
	callback    func(s *session, out io.Writer, args []string) error
}

type session struct {
	world    *world.World
	current  string
	api      *client.Client
	page     client.Page
	player   *game.Player
	foe      *game.Foe
	roll     game.Roll
	lootable bool
}

var errExit = errors.New("exit")

func commands() map[string]cliCommand {
	return map[string]cliCommand{
		"help":      {"help", "Show available commands", commandHelp},
		"exit":      {"exit", "Quit Grimhold", commandExit},
		"map":       {"map", "Show regions, their danger and their exits", commandMap},
		"look":      {"look", "Describe the current region", commandLook},
		"travel":    {"travel", "travel <region>: go to a neighboring region", commandTravel},
		"bestiary":  {"bestiary", "List Open5e monsters (first page)", commandBestiary},
		"next":      {"next", "Next bestiary page", commandNext},
		"prev":      {"prev", "Previous bestiary page", commandPrev},
		"inspect":   {"inspect", "inspect <monster>: show a monster's stat block", commandInspect},
		"explore":   {"explore", "Search the region for a monster", commandExplore},
		"fight":     {"fight", "Attack the encountered monster", commandFight},
		"flee":      {"flee", "Try to flee the fight", commandFlee},
		"status":    {"status", "Show HP, XP, gold and kills", commandStatus},
		"character": {"character", "Show the character sheet", commandCharacter},
		"rest":      {"rest", "Full heal, once per region visit", commandRest},
		"save":      {"save", "Save the game to save.json", commandSave},
		"load":      {"load", "Load the game from save.json", commandLoad},
		"loot":      {"loot", "Search a defeated monster for a weapon", commandLoot},
		"equip":     {"equip", "equip <weapon>: wield a weapon from your bag", commandEquip},
	}
}

func Run(w *world.World, api *client.Client, in io.Reader, out io.Writer, resume bool) {
	s := &session{world: w, current: w.Start, api: api, player: game.NewPlayer(),
		roll: func(n int) int { return rand.IntN(n) + 1 }}
	cmds := commands()
	if resume {
		if err := commandLoad(s, out, nil); err != nil {
			fmt.Fprintln(out, "Error:", err)
		}
	}
	sc := bufio.NewScanner(in)
	for {
		fmt.Fprint(out, "grimhold > ")
		if !sc.Scan() {
			fmt.Fprintln(out)
			if err := sc.Err(); err != nil {
				fmt.Fprintln(out, "Read error:", err)
			}
			return
		}
		words := strings.Fields(strings.ToLower(sc.Text()))
		if len(words) == 0 {
			fmt.Fprintln(out, "Type a command (help for the list).")
			continue
		}
		cmd, ok := cmds[words[0]]
		if !ok {
			fmt.Fprintf(out, "Unknown command: %q. Type help.\n", words[0])
			continue
		}
		if err := cmd.callback(s, out, words[1:]); errors.Is(err, errExit) {
			return
		} else if err != nil {
			fmt.Fprintln(out, "Error:", err)
		}
	}
}

func noArgs(args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("this command takes no arguments (got %d)", len(args))
	}
	return nil
}

func commandHelp(_ *session, out io.Writer, args []string) error {
	if err := noArgs(args); err != nil {
		return err
	}
	cmds := commands()
	names := make([]string, 0, len(cmds))
	for n := range cmds {
		names = append(names, n)
	}
	sort.Strings(names)
	fmt.Fprintln(out, "Commands:")
	for _, n := range names {
		fmt.Fprintf(out, "  %-10s %s\n", n, cmds[n].description)
	}
	return nil
}

func commandExit(_ *session, out io.Writer, args []string) error {
	if err := noArgs(args); err != nil {
		return err
	}
	fmt.Fprintln(out, "Farewell, adventurer.")
	return errExit
}

func commandMap(s *session, out io.Writer, args []string) error {
	if err := noArgs(args); err != nil {
		return err
	}
	for _, r := range s.world.Regions {
		marker := " "
		if r.ID == s.current {
			marker = "*"
		}
		fmt.Fprintf(out, "%s %-18s danger %d  -> %s\n", marker, r.ID, r.Danger, strings.Join(r.Exits, ", "))
	}
	return nil
}

func commandLook(s *session, out io.Writer, args []string) error {
	if err := noArgs(args); err != nil {
		return err
	}
	r, _ := s.world.Region(s.current)
	fmt.Fprintf(out, "== %s == (danger %d)\n%s\nExits: %s\n", r.Name, r.Danger, r.Description, strings.Join(r.Exits, ", "))
	return nil
}

func commandTravel(s *session, out io.Writer, args []string) error {
	if len(args) != 1 {
		return errors.New("usage: travel <region>")
	}
	if s.foe != nil {
		return errors.New("not possible mid-fight: fight or flee")
	}
	if _, ok := s.world.Region(args[0]); !ok {
		return fmt.Errorf("unknown region: %q", args[0])
	}
	if here, _ := s.world.Region(s.current); !here.HasExit(args[0]) {
		return fmt.Errorf("no direct path from %q to %q", s.current, args[0])
	}
	s.current = args[0]
	s.player.Rested = false
	return commandLook(s, out, nil)
}

func commandBestiary(s *session, out io.Writer, args []string) error {
	if err := noArgs(args); err != nil {
		return err
	}
	return showPage(s, out, "")
}

func commandNext(s *session, out io.Writer, args []string) error {
	if err := noArgs(args); err != nil {
		return err
	}
	if s.page.Next == nil {
		return errors.New("no next page (type bestiary to start)")
	}
	return showPage(s, out, *s.page.Next)
}

func commandPrev(s *session, out io.Writer, args []string) error {
	if err := noArgs(args); err != nil {
		return err
	}
	if s.page.Previous == nil {
		return errors.New("no previous page")
	}
	return showPage(s, out, *s.page.Previous)
}

func showPage(s *session, out io.Writer, pageURL string) error {
	p, err := s.api.Monsters(pageURL)
	if err != nil {
		return err
	}
	s.page = p
	for _, m := range p.Results {
		fmt.Fprintf(out, "  %-30s CR %s\n", m.Slug, m.ChallengeRating)
	}
	fmt.Fprintf(out, "%d monsters in total. next / prev to browse, inspect <monster> for details.\n", p.Count)
	return nil
}

func commandInspect(s *session, out io.Writer, args []string) error {
	if len(args) == 0 {
		return errors.New("usage: inspect <monster>")
	}
	slug := strings.Join(args, "-")
	m, err := s.api.Monster(slug)
	if err != nil {
		return fmt.Errorf("monster %q: %w", slug, err)
	}
	fmt.Fprintf(out, "== %s ==\nHP %d  AC %d  CR %s\n", m.Name, m.HitPoints, m.ArmorClass, m.ChallengeRating)
	if len(m.Actions) > 0 {
		fmt.Fprintln(out, "Actions:")
		for _, a := range m.Actions {
			fmt.Fprintf(out, "  - %s: %s\n", a.Name, a.Desc)
		}
	}
	return nil
}

func commandExplore(s *session, out io.Writer, args []string) error {
	if err := noArgs(args); err != nil {
		return err
	}
	if s.foe != nil {
		return fmt.Errorf("you are already fighting %s: fight or flee", s.foe.Name)
	}
	r, _ := s.world.Region(s.current)
	if len(r.Foes) == 0 {
		fmt.Fprintln(out, "Nothing lurks here.")
		return nil
	}

	var worst client.Monster
	fetched := map[string]bool{}
	for range max(r.Danger, 1) {
		slug := r.Foes[s.roll(len(r.Foes))-1]
		if fetched[slug] {
			continue
		}
		fetched[slug] = true
		m, err := s.api.Monster(slug)
		if err != nil {
			return fmt.Errorf("monster %q: %w", slug, err)
		}
		if worst.Slug == "" || m.CR > worst.CR {
			worst = m
		}
	}
	s.foe = game.NewFoe(worst)
	s.lootable = false
	fmt.Fprintf(out, "A %s appears! (HP %d, AC %d, CR %s) fight or flee?\n", s.foe.Name, s.foe.HP, s.foe.AC, worst.ChallengeRating)
	return nil
}

func commandFight(s *session, out io.Writer, args []string) error {
	if err := noArgs(args); err != nil {
		return err
	}
	if s.foe == nil {
		return errors.New("no opponent (type explore)")
	}
	printLines(out, game.Round(s.player, s.foe, s.roll))
	if s.foe.HP <= 0 {
		xp, gold, levels := game.Victory(s.player, s.foe, s.roll)
		fmt.Fprintf(out, "%s is defeated! +%d XP, +%d gold.\n", s.foe.Name, xp, gold)
		if levels > 0 {
			fmt.Fprintf(out, "Level %d reached! Max HP %d, attack bonus +%d.\n", s.player.Level, s.player.MaxHP, s.player.Attack)
		}
		s.foe = nil
		s.lootable = true
		fmt.Fprintln(out, "Type loot to search the remains.")
		return nil
	}
	checkDefeat(s, out)
	return nil
}

func commandLoot(s *session, out io.Writer, args []string) error {
	if err := noArgs(args); err != nil {
		return err
	}
	if !s.lootable {
		return errors.New("nothing to loot: defeat a monster first")
	}
	ws, err := s.api.Weapons()
	if err != nil {
		return err
	}
	w, ok := game.Loot(s.player, ws, s.roll)
	s.lootable = false
	if !ok {
		fmt.Fprintln(out, "You find nothing of use.")
		return nil
	}
	fmt.Fprintf(out, "You find a %s (%s). equip %s to wield it.\n", w.Name, w.Dice, strings.ToLower(w.Name))
	return nil
}

func commandEquip(s *session, out io.Writer, args []string) error {
	if len(args) == 0 {
		return errors.New("usage: equip <weapon>")
	}
	name := strings.Join(args, " ")
	w, ok := s.player.Equip(name)
	if !ok {
		return fmt.Errorf("no %q in your bag", name)
	}
	fmt.Fprintf(out, "You wield the %s (%s+2).\n", w.Name, w.Dice)
	return nil
}

func commandFlee(s *session, out io.Writer, args []string) error {
	if err := noArgs(args); err != nil {
		return err
	}
	if s.foe == nil {
		return errors.New("nothing to flee from")
	}
	escaped, log := game.Flee(s.player, s.foe, s.roll)
	printLines(out, log)
	if escaped {
		s.foe = nil
		return nil
	}
	checkDefeat(s, out)
	return nil
}

func checkDefeat(s *session, out io.Writer) {
	if s.player.HP > 0 {
		return
	}
	game.Defeat(s.player)
	s.foe = nil
	s.current = s.world.Start
	s.player.Rested = false
	fmt.Fprintln(out, "You fall... and wake up at the crossroads with 1 HP.")
}

func printLines(out io.Writer, lines []string) {
	for _, l := range lines {
		fmt.Fprintln(out, l)
	}
}

func commandStatus(s *session, out io.Writer, args []string) error {
	if err := noArgs(args); err != nil {
		return err
	}
	p := s.player
	fmt.Fprintf(out, "HP %d/%d  AC %d  XP %d  Gold %d\n", p.HP, p.MaxHP, p.AC, p.XP, p.Gold)
	if len(p.Kills) == 0 {
		return nil
	}
	names := make([]string, 0, len(p.Kills))
	for n := range p.Kills {
		names = append(names, n)
	}
	sort.Strings(names)
	fmt.Fprintln(out, "Kills:")
	for _, n := range names {
		fmt.Fprintf(out, "  %-20s x%d\n", n, p.Kills[n])
	}
	return nil
}

func commandCharacter(s *session, out io.Writer, args []string) error {
	if err := noArgs(args); err != nil {
		return err
	}
	p := s.player
	fmt.Fprintf(out, "== Character ==\nLevel %d  XP %d/%d\nHP %d/%d  AC %d  Attack +%d\nWeapon: %s (%s+2)\nGold %d\nBag:\n",
		p.Level, p.XP, p.NextLevelXP(), p.HP, p.MaxHP, p.AC, p.Attack, p.Weapon, p.Dice, p.Gold)
	for _, w := range p.Bag {
		fmt.Fprintf(out, "  %-20s %s\n", w.Name, w.Dice)
	}
	return nil
}

func commandRest(s *session, out io.Writer, args []string) error {
	if err := noArgs(args); err != nil {
		return err
	}
	if s.foe != nil {
		return errors.New("not possible mid-fight: fight or flee")
	}
	if !s.player.Rest() {
		return errors.New("you have already rested here: travel before resting again")
	}
	fmt.Fprintf(out, "You rest. HP %d/%d.\n", s.player.HP, s.player.MaxHP)
	return nil
}
