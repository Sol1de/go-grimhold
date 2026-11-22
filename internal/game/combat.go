package game

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Sol1de/go-grimhold/internal/client"
)

type Roll func(sides int) int

type Player struct {
	HP, MaxHP, AC, Attack, XP, Gold, Level int
	Weapon                                 string
	Rested                                 bool
	Kills                                  map[string]int
}

func NewPlayer() *Player {
	return &Player{HP: 20, MaxHP: 20, AC: 13, Attack: 4, Level: 1, Weapon: "Longsword (1d8+2)", Kills: map[string]int{}}
}

func (p *Player) NextLevelXP() int { return p.Level * p.Level * 50 }

func (p *Player) gainXP(xp int) int {
	p.XP += xp
	gained := 0
	for p.XP >= p.NextLevelXP() {
		p.Level++
		p.MaxHP += 5
		p.HP += 5
		p.Attack++
		gained++
	}
	return gained
}

func (p *Player) Rest() bool {
	if p.Rested {
		return false
	}
	p.HP, p.Rested = p.MaxHP, true
	return true
}

type Foe struct {
	Name                     string
	HP, AC, Attack, DmgBonus int
	DmgN, DmgSides           int
	CR                       float64
}

func NewFoe(m client.Monster) *Foe {
	f := &Foe{Name: m.Name, HP: m.HitPoints, AC: m.ArmorClass, CR: m.CR, Attack: 2, DmgN: 1, DmgSides: 6}
	for _, a := range m.Actions {
		if n, s, ok := parseDice(a.DamageDice); ok {
			f.Attack, f.DmgN, f.DmgSides, f.DmgBonus = a.AttackBonus, n, s, a.DamageBonus
			break
		}
	}
	return f
}

func parseDice(s string) (n, sides int, ok bool) {
	a, b, found := strings.Cut(strings.TrimSpace(s), "d")
	n, err1 := strconv.Atoi(a)
	sides, err2 := strconv.Atoi(b)
	if !found || err1 != nil || err2 != nil || n < 1 || sides < 1 {
		return 0, 0, false
	}
	return n, sides, true
}

func sum(roll Roll, n, sides, bonus int) int {
	t := bonus
	for range n {
		t += roll(sides)
	}
	return max(t, 1)
}

func Round(p *Player, f *Foe, roll Roll) []string {
	var log []string
	if d20 := roll(20); d20+p.Attack >= f.AC {
		dmg := sum(roll, 1, 8, 2)
		f.HP -= dmg
		log = append(log, fmt.Sprintf("You hit %s (%d+%d vs AC %d): %d damage. It has %d HP left.", f.Name, d20, p.Attack, f.AC, dmg, max(f.HP, 0)))
	} else {
		log = append(log, fmt.Sprintf("You miss %s (%d+%d vs AC %d).", f.Name, d20, p.Attack, f.AC))
	}
	if f.HP > 0 {
		log = append(log, foeAttack(p, f, roll))
	}
	return log
}

func Flee(p *Player, f *Foe, roll Roll) (bool, []string) {
	if d20 := roll(20); d20 >= 10 {
		return true, []string{fmt.Sprintf("You flee (%d).", d20)}
	}
	return false, []string{"You stumble while fleeing!", foeAttack(p, f, roll)}
}

func foeAttack(p *Player, f *Foe, roll Roll) string {
	d20 := roll(20)
	if d20+f.Attack < p.AC {
		return fmt.Sprintf("%s misses you (%d+%d vs AC %d).", f.Name, d20, f.Attack, p.AC)
	}
	dmg := sum(roll, f.DmgN, f.DmgSides, f.DmgBonus)
	p.HP -= dmg
	return fmt.Sprintf("%s hits you: %d damage. You have %d HP left.", f.Name, dmg, max(p.HP, 0))
}

var xpByCR = map[float64]int{
	0: 10, 0.125: 25, 0.25: 50, 0.5: 100, 1: 200, 2: 450, 3: 700, 4: 1100, 5: 1800,
	6: 2300, 7: 2900, 8: 3900, 9: 5000, 10: 5900, 11: 7200, 12: 8400, 13: 10000,
	14: 11500, 15: 13000, 16: 15000, 17: 18000, 18: 20000, 19: 22000, 20: 25000,
}

func Victory(p *Player, f *Foe, roll Roll) (xp, gold, levels int) {
	xp = xpByCR[f.CR]
	gold = roll(6) * max(1, int(f.CR*2))
	levels = p.gainXP(xp)
	p.Gold += gold
	p.Kills[f.Name]++
	return xp, gold, levels
}

func Defeat(p *Player) { p.HP = 1 }
