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
	world   *world.World
	current string
	api     *client.Client
	page    client.Page
	player  *game.Player
	foe     *game.Foe
	roll    game.Roll
}

var errExit = errors.New("exit")

func commands() map[string]cliCommand {
	return map[string]cliCommand{
		"help":     {"help", "Affiche les commandes disponibles", commandHelp},
		"exit":     {"exit", "Quitte Grimhold", commandExit},
		"map":      {"map", "Affiche les régions, leur danger et leurs sorties", commandMap},
		"look":     {"look", "Décrit la région actuelle", commandLook},
		"travel":   {"travel", "travel <region> : se rend dans une région voisine", commandTravel},
		"bestiary": {"bestiary", "Liste les monstres d'Open5e (première page)", commandBestiary},
		"next":     {"next", "Page suivante du bestiaire", commandNext},
		"prev":     {"prev", "Page précédente du bestiaire", commandPrev},
		"inspect":  {"inspect", "inspect <monstre> : fiche de stats d'un monstre", commandInspect},
		"explore":  {"explore", "Explore la région à la recherche d'un monstre", commandExplore},
		"fight":    {"fight", "Attaque le monstre rencontré", commandFight},
		"flee":     {"flee", "Tente de fuir le combat", commandFlee},
		"status":   {"status", "Affiche PV, XP, or et victimes", commandStatus},
	}
}

func Run(w *world.World, api *client.Client, in io.Reader, out io.Writer) {
	s := &session{world: w, current: w.Start, api: api, player: game.NewPlayer(),
		roll: func(n int) int { return rand.IntN(n) + 1 }}
	cmds := commands()
	sc := bufio.NewScanner(in)
	for {
		fmt.Fprint(out, "grimhold > ")
		if !sc.Scan() {
			fmt.Fprintln(out)
			if err := sc.Err(); err != nil {
				fmt.Fprintln(out, "Erreur de lecture :", err)
			}
			return
		}
		words := strings.Fields(strings.ToLower(sc.Text()))
		if len(words) == 0 {
			fmt.Fprintln(out, "Tapez une commande (help pour la liste).")
			continue
		}
		cmd, ok := cmds[words[0]]
		if !ok {
			fmt.Fprintf(out, "Commande inconnue : %q. Tapez help.\n", words[0])
			continue
		}
		if err := cmd.callback(s, out, words[1:]); errors.Is(err, errExit) {
			return
		} else if err != nil {
			fmt.Fprintln(out, "Erreur :", err)
		}
	}
}

func noArgs(args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("cette commande ne prend pas d'argument (reçu %d)", len(args))
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
	fmt.Fprintln(out, "Commandes :")
	for _, n := range names {
		fmt.Fprintf(out, "  %-10s %s\n", n, cmds[n].description)
	}
	return nil
}

func commandExit(_ *session, out io.Writer, args []string) error {
	if err := noArgs(args); err != nil {
		return err
	}
	fmt.Fprintln(out, "Adieu, aventurier.")
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
	fmt.Fprintf(out, "== %s == (danger %d)\n%s\nSorties : %s\n", r.Name, r.Danger, r.Description, strings.Join(r.Exits, ", "))
	return nil
}

func commandTravel(s *session, out io.Writer, args []string) error {
	if len(args) != 1 {
		return errors.New("usage : travel <region>")
	}
	if s.foe != nil {
		return errors.New("impossible en plein combat : fight ou flee")
	}
	if _, ok := s.world.Region(args[0]); !ok {
		return fmt.Errorf("région inconnue : %q", args[0])
	}
	if here, _ := s.world.Region(s.current); !here.HasExit(args[0]) {
		return fmt.Errorf("aucun chemin direct de %q vers %q", s.current, args[0])
	}
	s.current = args[0]
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
		return errors.New("pas de page suivante (tapez bestiary pour commencer)")
	}
	return showPage(s, out, *s.page.Next)
}

func commandPrev(s *session, out io.Writer, args []string) error {
	if err := noArgs(args); err != nil {
		return err
	}
	if s.page.Previous == nil {
		return errors.New("pas de page précédente")
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
	fmt.Fprintf(out, "%d monstres au total. next / prev pour naviguer, inspect <monstre> pour les détails.\n", p.Count)
	return nil
}

func commandInspect(s *session, out io.Writer, args []string) error {
	if len(args) == 0 {
		return errors.New("usage : inspect <monstre>")
	}
	slug := strings.Join(args, "-")
	m, err := s.api.Monster(slug)
	if err != nil {
		return fmt.Errorf("monstre %q : %w", slug, err)
	}
	fmt.Fprintf(out, "== %s ==\nPV %d  CA %d  CR %s\n", m.Name, m.HitPoints, m.ArmorClass, m.ChallengeRating)
	if len(m.Actions) > 0 {
		fmt.Fprintln(out, "Actions :")
		for _, a := range m.Actions {
			fmt.Fprintf(out, "  - %s : %s\n", a.Name, a.Desc)
		}
	}
	return nil
}

func commandExplore(s *session, out io.Writer, args []string) error {
	if err := noArgs(args); err != nil {
		return err
	}
	if s.foe != nil {
		return fmt.Errorf("vous combattez déjà %s : fight ou flee", s.foe.Name)
	}
	r, _ := s.world.Region(s.current)
	if len(r.Foes) == 0 {
		fmt.Fprintln(out, "Rien ne rôde ici.")
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
			return fmt.Errorf("monstre %q : %w", slug, err)
		}
		if worst.Slug == "" || m.CR > worst.CR {
			worst = m
		}
	}
	s.foe = game.NewFoe(worst)
	fmt.Fprintf(out, "Un %s surgit ! (PV %d, CA %d, CR %s) fight ou flee ?\n", s.foe.Name, s.foe.HP, s.foe.AC, worst.ChallengeRating)
	return nil
}

func commandFight(s *session, out io.Writer, args []string) error {
	if err := noArgs(args); err != nil {
		return err
	}
	if s.foe == nil {
		return errors.New("aucun adversaire (tapez explore)")
	}
	printLines(out, game.Round(s.player, s.foe, s.roll))
	if s.foe.HP <= 0 {
		xp, gold := game.Victory(s.player, s.foe, s.roll)
		fmt.Fprintf(out, "%s est vaincu ! +%d XP, +%d or.\n", s.foe.Name, xp, gold)
		s.foe = nil
		return nil
	}
	checkDefeat(s, out)
	return nil
}

func commandFlee(s *session, out io.Writer, args []string) error {
	if err := noArgs(args); err != nil {
		return err
	}
	if s.foe == nil {
		return errors.New("rien à fuir")
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
	fmt.Fprintln(out, "Vous tombez... et vous réveillez au carrefour avec 1 PV.")
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
	fmt.Fprintf(out, "PV %d/%d  CA %d  XP %d  Or %d\n", p.HP, p.MaxHP, p.AC, p.XP, p.Gold)
	if len(p.Kills) == 0 {
		return nil
	}
	names := make([]string, 0, len(p.Kills))
	for n := range p.Kills {
		names = append(names, n)
	}
	sort.Strings(names)
	fmt.Fprintln(out, "Victimes :")
	for _, n := range names {
		fmt.Fprintf(out, "  %-20s x%d\n", n, p.Kills[n])
	}
	return nil
}
