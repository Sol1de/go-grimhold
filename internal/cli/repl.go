package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"

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
}

var errExit = errors.New("exit")

func commands() map[string]cliCommand {
	return map[string]cliCommand{
		"help":   {"help", "Affiche les commandes disponibles", commandHelp},
		"exit":   {"exit", "Quitte Grimhold", commandExit},
		"map":    {"map", "Affiche les régions, leur danger et leurs sorties", commandMap},
		"look":   {"look", "Décrit la région actuelle", commandLook},
		"travel": {"travel", "travel <region> : se rend dans une région voisine", commandTravel},
	}
}

func Run(w *world.World, in io.Reader, out io.Writer) {
	s := &session{world: w, current: w.Start}
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
	if _, ok := s.world.Region(args[0]); !ok {
		return fmt.Errorf("région inconnue : %q", args[0])
	}
	if here, _ := s.world.Region(s.current); !here.HasExit(args[0]) {
		return fmt.Errorf("aucun chemin direct de %q vers %q", s.current, args[0])
	}
	s.current = args[0]
	return commandLook(s, out, nil)
}
