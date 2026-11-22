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
		return errors.New("not possible mid-fight: fight or flee")
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
	fmt.Fprintf(out, "Game saved to %s.\n", saveFile)
	return nil
}

func commandLoad(s *session, out io.Writer, args []string) error {
	if err := noArgs(args); err != nil {
		return err
	}
	b, err := os.ReadFile(saveFile)
	if errors.Is(err, fs.ErrNotExist) {
		fmt.Fprintln(out, "No save found.")
		return nil
	}
	if err != nil {
		return err
	}
	var d saveData
	if err := json.Unmarshal(b, &d); err != nil {
		return fmt.Errorf("unreadable save: %w", err)
	}
	if _, ok := s.world.Region(d.Current); !ok || d.Player == nil {
		return errors.New("invalid save")
	}
	if d.Player.Kills == nil {
		d.Player.Kills = map[string]int{}
	}
	if d.Player.Dice == "" {
		d.Player.Weapon, d.Player.Dice, d.Player.Bag = game.Longsword.Name, game.Longsword.Dice, []game.Weapon{game.Longsword}
	}
	s.current, s.player, s.foe, s.lootable = d.Current, d.Player, nil, false
	fmt.Fprintln(out, "Game loaded.")
	return commandLook(s, out, nil)
}
