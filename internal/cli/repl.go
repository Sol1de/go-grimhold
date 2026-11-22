package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
)

type cliCommand struct {
	name        string
	description string
	callback    func(out io.Writer, args []string) error
}

var errExit = errors.New("exit")

func commands() map[string]cliCommand {
	return map[string]cliCommand{
		"help": {"help", "Affiche les commandes disponibles", commandHelp},
		"exit": {"exit", "Quitte Grimhold", commandExit},
	}
}

func Run(in io.Reader, out io.Writer) {
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
		if err := cmd.callback(out, words[1:]); errors.Is(err, errExit) {
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

func commandHelp(out io.Writer, args []string) error {
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

func commandExit(out io.Writer, args []string) error {
	if err := noArgs(args); err != nil {
		return err
	}
	fmt.Fprintln(out, "Adieu, aventurier.")
	return errExit
}
