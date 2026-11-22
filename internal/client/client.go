package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const BaseURL = "https://api.open5e.com/v1/monsters/"

const WeaponsURL = "https://api.open5e.com/v1/weapons/?limit=100"

type Weapon struct {
	Name       string `json:"name"`
	DamageDice string `json:"damage_dice"`
}

type Action struct {
	Name        string `json:"name"`
	Desc        string `json:"desc"`
	AttackBonus int    `json:"attack_bonus"`
	DamageDice  string `json:"damage_dice"`
	DamageBonus int    `json:"damage_bonus"`
}

type Monster struct {
	Slug            string   `json:"slug"`
	Name            string   `json:"name"`
	HitPoints       int      `json:"hit_points"`
	ArmorClass      int      `json:"armor_class"`
	ChallengeRating string   `json:"challenge_rating"`
	CR              float64  `json:"cr"`
	Actions         []Action `json:"actions"`
}

type Page struct {
	Count    int       `json:"count"`
	Next     *string   `json:"next"`
	Previous *string   `json:"previous"`
	Results  []Monster `json:"results"`
}

type Client struct {
	BaseURL string
	HTTP    *http.Client
}

func New() *Client {
	return &Client{BaseURL: BaseURL, HTTP: &http.Client{Timeout: 10 * time.Second}}
}

func (c *Client) Monsters(pageURL string) (Page, error) {
	if pageURL == "" {
		pageURL = c.BaseURL
	}
	var p Page
	return p, c.get(pageURL, &p)
}

func (c *Client) Monster(slug string) (Monster, error) {
	var m Monster
	return m, c.get(c.BaseURL+url.PathEscape(slug)+"/", &m)
}

func (c *Client) Weapons() ([]Weapon, error) {
	var p struct {
		Results []Weapon `json:"results"`
	}
	return p.Results, c.get(WeaponsURL, &p)
}

func (c *Client) get(u string, v any) error {
	resp, err := c.HTTP.Get(u)
	if err != nil {
		return fmt.Errorf("Open5e API unreachable: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("not found")
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Open5e API: status %s", resp.Status)
	}
	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return fmt.Errorf("invalid Open5e response: %w", err)
	}
	return nil
}
