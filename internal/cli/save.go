package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/Sol1de/go-grimhold/internal/game"
)

const saveFile = "save.json"

type saveData struct {
	Current string       `json:"current"`
	Player  *game.Player `json:"player"`
}

func commandSave(s *session, out io.Writer, args []string) error {
	if err := noArgs(args); err != nil {
		return err
	}
	if s.foe != nil {
		return errors.New("impossible en plein combat : fight ou flee")
	}
	b, err := json.MarshalIndent(saveData{s.current, s.player}, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(saveFile+".tmp", b, 0o644); err != nil {
		return err
	}
	if err := os.Rename(saveFile+".tmp", saveFile); err != nil {
		return err
	}
	fmt.Fprintf(out, "Partie sauvegardée dans %s.\n", saveFile)
	return nil
}

func commandLoad(s *session, out io.Writer, args []string) error {
	if err := noArgs(args); err != nil {
		return err
	}
	b, err := os.ReadFile(saveFile)
	if errors.Is(err, fs.ErrNotExist) {
		fmt.Fprintln(out, "Aucune sauvegarde trouvée.")
		return nil
	}
	if err != nil {
		return err
	}
	var d saveData
	if err := json.Unmarshal(b, &d); err != nil {
		return fmt.Errorf("sauvegarde illisible : %w", err)
	}
	if _, ok := s.world.Region(d.Current); !ok || d.Player == nil {
		return errors.New("sauvegarde invalide")
	}
	if d.Player.Kills == nil {
		d.Player.Kills = map[string]int{}
	}
	s.current, s.player, s.foe = d.Current, d.Player, nil
	fmt.Fprintln(out, "Partie chargée.")
	return commandLook(s, out, nil)
}
