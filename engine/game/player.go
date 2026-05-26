package game

type Player int

type Team int // Red/Yellow: 1, Blue/Green: -1.

var (
	RedYellow = []Player{0, 2}
	BlueGreen = []Player{1, 3}
	Opponents = [2][]Player{
		BlueGreen,
		RedYellow,
	}
)

// IsTeamMate returns true if p and other are on the same team (including p == other).
func (p Player) IsTeamMate(other Player) bool {
	return (p^other)&1 == 0 // Last bit must match.
}

// Team returns 1 for Red/Yellow and -1 for Blue/Green.
func (p Player) Team() Team {
	t := p & 1
	return Team(1 - 2*t)
}

// Teammates returns the two players on the same team.
func (p Player) Teammates() []Player {
	return Opponents[(p+1)%2]
}

// Opponents returns the two players on the opposite team.
func (p Player) Opponents() []Player {
	return Opponents[p%2]
}

// Opposite returns the opposite team.
func (t Team) Opposite() Team {
	return t * -1
}

func (p Player) String() string {
	switch p {
	case 0:
		return "Red"
	case 1:
		return "Blue"
	case 2:
		return "Yellow"
	case 3:
		return "Green"
	default:
		panic("unsupported player")
	}
}

// String implements the Stringer interface.
func (t Team) String() string {
	switch t {
	case 1:
		return "Red/Yellow"
	case -1:
		return "Blue/Green"
	default:
		panic("unsupported team")
	}
}
