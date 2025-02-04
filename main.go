package main

import (
	"fmt"
	"math"
	"strings"
	"time"
)

type DisplayCharacters int

const (
	FULL DisplayCharacters = iota
	DARK
	MEDIUM
	LIGHT
	EMPTY
)

/*
* █
* ▓
* ▒
* ░
*
 */

func (dc DisplayCharacters) value() rune {
	switch dc {
	case FULL:
		return '█'
	case DARK:
		return '▓'
	case MEDIUM:
		return '▒'
	case LIGHT:
		return '░'
	case EMPTY:
		return ' '
	default:
		return ' '
	}
}

type Point struct {
	a float64
	b float64
}

type ReactionDiffusion struct {
	DIFFUSION_A  float64
	DIFFUSION_B  float64
	FEED         float64
	KILL         float64
	FILL_AMOUNT  float64
	Grid         [][]Point
	Next         [][]Point
	Frame        [][]rune
	height       int
	width        int
	centerHeight int
	centerWidth  int
	radius       int
}

func NewReactionDiffusion(height int, width int, infillRatio float64) *ReactionDiffusion {
	rd := &ReactionDiffusion{
		DIFFUSION_A:  1.0,
		DIFFUSION_B:  0.5,
		FEED:         0.055,
		KILL:         0.062,
		FILL_AMOUNT:  0.1,
		height:       height,
		width:        width,
		centerHeight: height / 2,
		centerWidth:  width / 2,
		radius:       int(float64(width/2) * infillRatio),
	}

	// Initialize the Grid and Next slices
	rd.Grid = make([][]Point, height)
	rd.Next = make([][]Point, height)
	rd.Frame = make([][]rune, height)
	for i := range rd.Grid {
		rd.Grid[i] = make([]Point, width)
		rd.Next[i] = make([]Point, width)
		rd.Frame[i] = make([]rune, width)
		for j := range rd.Grid[i] {
			rd.Grid[i][j] = Point{a: 1.0, b: 0.0}
			rd.Next[i][j] = Point{a: 1.0, b: 0.0}
			rd.Frame[i][j] = EMPTY.value()
		}
	}

	// Init block of radius in center
	for row := rd.centerWidth - rd.radius; row < (rd.centerWidth + rd.radius); row++ {
		for item := rd.centerHeight - rd.radius; item < (rd.centerHeight + rd.radius); item++ {
			rd.Grid[row][item].b = 1.0
		}
	}

	return rd
}

func (rd *ReactionDiffusion) laplace_a(x int, y int) float64 {
	sum_a := 0.0

	sum_a += rd.Grid[x][y].a * -1.0
	sum_a += rd.Grid[x-1][y].a * 0.2
	sum_a += rd.Grid[x+1][y].a * 0.2
	sum_a += rd.Grid[x][y+1].a * 0.2
	sum_a += rd.Grid[x][y-1].a * 0.2
	sum_a += rd.Grid[x-1][y-1].a * 0.05
	sum_a += rd.Grid[x+1][y-1].a * 0.05
	sum_a += rd.Grid[x+1][y+1].a * 0.05
	sum_a += rd.Grid[x-1][y+1].a * 0.05

	return sum_a
}

func (rd *ReactionDiffusion) laplace_b(x int, y int) float64 {
	sum_b := 0.0

	sum_b += rd.Grid[x][y].b * -1.0
	sum_b += rd.Grid[x-1][y].b * 0.2
	sum_b += rd.Grid[x+1][y].b * 0.2
	sum_b += rd.Grid[x][y+1].b * 0.2
	sum_b += rd.Grid[x][y-1].b * 0.2
	sum_b += rd.Grid[x-1][y-1].b * 0.05
	sum_b += rd.Grid[x+1][y-1].b * 0.05
	sum_b += rd.Grid[x+1][y+1].b * 0.05
	sum_b += rd.Grid[x-1][y+1].b * 0.05

	return sum_b
}

func (rd *ReactionDiffusion) process_grid() {
	for x := 1; x < (len(rd.Grid) - 1); x++ {
		for y := 1; y < (len(rd.Grid[x]) - 1); y++ {
			a := rd.Grid[x][y].a
			b := rd.Grid[x][y].b
			next_a := a + rd.DIFFUSION_A*rd.laplace_a(x, y) - a*b*b + rd.FEED*(1.0-a)
			next_b := b + rd.DIFFUSION_B*rd.laplace_b(x, y) + a*b*b - (rd.KILL+rd.FEED)*b

			rd.Next[x][y].a = constrainFloat64(next_a, 0.0, 1.0)
			rd.Next[x][y].b = constrainFloat64(next_b, 0.0, 1.0)
		}
	}
}
func (rd *ReactionDiffusion) GetNextFrame() string {
	rd.process_grid()

	for row := range len(rd.Grid) {
		for item := range len(rd.Grid[row]) {
			a := rd.Grid[row][item].a
			b := rd.Grid[row][item].b
			c := math.Floor((a - b) * 255.0)
			d := constrainInt(int(c), 0, 255)

			if d >= 0 && d <= 50 {
				rd.Frame[row][item] = FULL.value()
			} else if d >= 51 && d <= 101 {
				rd.Frame[row][item] = DARK.value()
			} else if d >= 102 && d <= 152 {
				rd.Frame[row][item] = MEDIUM.value()
			} else if d >= 153 && d <= 203 {
				rd.Frame[row][item] = LIGHT.value()
			} else if d >= 204 {
				rd.Frame[row][item] = EMPTY.value()
			}
		}
	}

	frame_string := strings.Builder{}
	for row := range len(rd.Frame) {
		for item := range len(rd.Frame[row]) {
			frame_string.WriteRune(rd.Frame[row][item])
		}
		frame_string.WriteRune('\n')
	}

	rd.swap()
	return frame_string.String()
}

func (rd *ReactionDiffusion) swap() {
	rd.Grid, rd.Next = rd.Next, rd.Grid
}

func constrainFloat64(value float64, min float64, max float64) float64 {
	if value < min {
		return min
	} else if value > max {
		return max
	} else {
		return value
	}
}
func constrainInt(value int, min int, max int) int {
	if value < min {
		return min
	} else if value > max {
		return max
	} else {
		return value
	}
}

func main() {
	rd := NewReactionDiffusion(150, 150, 0.1)
	i := 0
	for {
		frame := rd.GetNextFrame()

		fmt.Printf("%c[2J%c[1;1H", 27, 27)
		fmt.Printf("%s", frame)
		fmt.Printf("\n%d", i)
		i += 1
		time.Sleep(17 * time.Millisecond)
	}
}
