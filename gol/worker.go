package gol

func worker(world [][]byte, p Params, startY, endY int, out chan [][]byte) {

  // make a new slice that is the size of the part of the
  // world to be computed by this worker
	newworld := make([][]byte, endY - startY)
	for i := range newworld {
		newworld[i] = make([]byte, p.ImageWidth)
	}

	// check each cell in the appropriate portion of the world
  for i := range world[startY:endY] {

    // variable necessary to differentiate between local slice 
    // coordinates and global world coordinates
    worldIndex := startY + i

		for j := range world[worldIndex] {

			sum := 0

			// counts the number of alive neighbours
			for m := -1; m <= 1; m++ {
				for n := -1; n <= 1; n++ {

          // the value of the currently observed cell
          // should not be added to the sum
					if m != 0 || n != 0 {
						dy := (worldIndex + m + p.ImageHeight) % p.ImageHeight
						dx := (j + n + p.ImageWidth) % p.ImageWidth

            // byte values of 255 to 1 for proccessing
						sum += (int(world[dy][dx]) / 255)
					}
				}
			}

			// apply rules corresponding to the total surronding alive
      // cells in context for the state of the current cell
			if sum == 3 || (int(world[worldIndex][j])/255)+sum == 3 {
				newworld[i][j] = 255
			}
		}
	}
  // send computed data to the channel provided by the arguments
  out <- newworld
}


