package gol

import (
)

func worker(world [][]byte, p Params, startY, endY int) [][]byte {
	newworld := make([][]byte, p.ImageWidth)
	for i := range newworld {
		newworld[i] = make([]byte, endY - startY)
	}

	// check each cell in world
  for i := range world[startY:endY] {
		for j := range world[i] {

			sum := 0

			// alive neighbour count
			for m := -1; m <= 1; m++ {
				for n := -1; n <= 1; n++ {
					if m != 0 || n != 0 {
						dx := (i + m + p.ImageWidth) % p.ImageWidth
						dy := (j + n + p.ImageHeight) % p.ImageHeight

						sum += (int(world[dx][dy]) / 255)

					}
				}
			}

			//apply rules
			if sum == 3 || (int(world[i][j])/255)+sum == 3 {
				newworld[i % (startY - endY)][j] = 255
			}
		}
	}

	return newworld
}


